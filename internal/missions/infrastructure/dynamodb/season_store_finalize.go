package dynamodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// maxIntentsPerTransaction is DynamoDB's TransactWriteItems ceiling (100)
// minus the season write itself; a larger batch can never commit atomically,
// so the store rejects it instead of emitting a partial finalization.
const maxIntentsPerTransaction = 99

// placementFactPayload is the versioned contract of
// missions.season.placement-awarded v1.0 (snake_case).
type placementFactPayload struct {
	SeasonID    string   `json:"season_id"`
	Category    string   `json:"category"`
	Track       string   `json:"track"`
	RecipientID string   `json:"recipient_id"`
	Rank        int      `json:"rank"`
	FinalMerits int      `json:"final_merits"`
	MemberIDs   []string `json:"member_ids,omitempty"`
}

// placementEnvelope is the domain-event envelope written to the outbox.
type placementEnvelope struct {
	EventID       string               `json:"event_id"`
	EventType     string               `json:"event_type"`
	EventVersion  string               `json:"event_version"`
	Source        string               `json:"source"`
	OccurredAt    string               `json:"occurred_at"`
	CorrelationID string               `json:"correlation_id"`
	Payload       placementFactPayload `json:"payload"`
}

// SaveFinalized persists the finalized season and every placement-fact outbox
// intent in one DynamoDB transaction. The season write compare-and-swaps on
// State = FREEZING: a phantom season is never created, and a concurrent or
// repeated finalization can never overwrite the season or reset outbox rows
// that may already be published.
func (s *SeasonStore) SaveFinalized(season domain.Season, facts []domain.PlacementFact) error {
	if len(facts) > maxIntentsPerTransaction {
		return ErrOutboxBatchTooLarge
	}

	now := time.Now().UTC().Format(time.RFC3339)
	items := make([]types.TransactWriteItem, 0, len(facts)+1)
	items = append(items, types.TransactWriteItem{Put: &types.Put{
		TableName:           aws.String(s.tableName),
		Item:                seasonItem(season),
		ConditionExpression: aws.String("attribute_exists(PK) AND #state = :freezing"),
		ExpressionAttributeNames: map[string]string{
			"#state": "State",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":freezing": &types.AttributeValueMemberS{Value: string(domain.SeasonStateFreezing)},
		},
	}})
	for _, fact := range facts {
		item, err := outboxPlacementItem(fact, now)
		if err != nil {
			return err
		}
		items = append(items, types.TransactWriteItem{Put: &types.Put{
			TableName: aws.String(s.tableName),
			Item:      item,
		}})
	}
	_, err := s.client.Client().TransactWriteItems(context.Background(), &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	})
	if err != nil {
		var canceled *types.TransactionCanceledException
		if errors.As(err, &canceled) && transactionFailedOnCondition(canceled) {
			return s.finalizeConditionError(season.SeasonID)
		}
		return fmt.Errorf("failed to save finalized season with placement facts: %w", err)
	}
	return nil
}

// finalizeConditionError separates a missing season from a state-transition
// conflict after the compare-and-swap condition rejected the write.
func (s *SeasonStore) finalizeConditionError(seasonID string) error {
	if _, err := s.Get(seasonID); err != nil {
		if errors.Is(err, domain.ErrSeasonNotFound) {
			return domain.ErrSeasonNotFound
		}
		return fmt.Errorf("failed to verify season %s after rejected finalization: %w", seasonID, err)
	}
	return domain.ErrInvalidSeasonStateTransition
}

// transactionFailedOnCondition reports whether the transaction was canceled
// by a failed condition check.
func transactionFailedOnCondition(err *types.TransactionCanceledException) bool {
	for _, reason := range err.CancellationReasons {
		if aws.ToString(reason.Code) == "ConditionalCheckFailed" {
			return true
		}
	}
	return false
}

// outboxPlacementItem maps a placement fact to its pending outbox intent.
func outboxPlacementItem(fact domain.PlacementFact, occurredAt string) (map[string]types.AttributeValue, error) {
	envelope := placementEnvelope{
		EventID:       fact.EventID,
		EventType:     "missions.season.placement-awarded",
		EventVersion:  "1.0",
		Source:        application.EventSource,
		OccurredAt:    occurredAt,
		CorrelationID: outboxCorrelationID(fact.SeasonID),
		Payload: placementFactPayload{
			SeasonID:    fact.SeasonID,
			Category:    fact.Category,
			Track:       string(fact.Track),
			RecipientID: fact.RecipientID,
			Rank:        fact.Rank,
			FinalMerits: fact.FinalMerits,
			MemberIDs:   fact.MemberIDs,
		},
	}
	raw, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal placement fact %s: %w", fact.EventID, err)
	}
	return map[string]types.AttributeValue{
		"PK":        &types.AttributeValueMemberS{Value: outboxPartition},
		"SK":        &types.AttributeValueMemberS{Value: outboxSortKey(fact.EventID)},
		"Status":    &types.AttributeValueMemberS{Value: outboxStatusPending},
		"EventID":   &types.AttributeValueMemberS{Value: fact.EventID},
		"EventType": &types.AttributeValueMemberS{Value: envelope.EventType},
		"Envelope":  &types.AttributeValueMemberS{Value: string(raw)},
	}, nil
}

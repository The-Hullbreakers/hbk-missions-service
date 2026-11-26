package dynamodb

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// PolicyStore persists runtime-tunable policies; a missing document reads as
// the shipped default so first-run behaviour is unchanged.
type PolicyStore struct {
	client    *DynamoDBClient
	tableName string
}

// NewPolicyStore creates a DynamoDB policy store.
func NewPolicyStore(client *DynamoDBClient, tableName string) *PolicyStore {
	return &PolicyStore{client: client, tableName: tableName}
}

const activeLimitSortKey = "ACTIVE_LIMIT"

// GetActiveLimit reads the persisted active-contract limit; absent document
// falls back to the shipped default of 5.
func (s *PolicyStore) GetActiveLimit() (domain.ActiveLimitPolicy, error) {
	ctx := context.Background()

	result, err := s.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "POLICY"},
			"SK": &types.AttributeValueMemberS{Value: activeLimitSortKey},
		},
	})
	if err != nil {
		return domain.ActiveLimitPolicy{}, fmt.Errorf("failed to read policy: %w", err)
	}
	if result.Item == nil {
		return domain.NewActiveLimitPolicy(5, nil), nil
	}

	overrides := map[string]int{}
	for category, raw := range getStringMap(result.Item, "Overrides") {
		if n, convErr := strconv.Atoi(raw); convErr == nil {
			overrides[category] = n
		}
	}
	return domain.NewActiveLimitPolicy(getInt(result.Item, "DefaultLimit"), overrides), nil
}

// SaveActiveLimit persists the active-contract limit document.
func (s *PolicyStore) SaveActiveLimit(policy domain.ActiveLimitPolicy) error {
	ctx := context.Background()

	attrs := map[string]types.AttributeValue{
		"PK":           &types.AttributeValueMemberS{Value: "POLICY"},
		"SK":           &types.AttributeValueMemberS{Value: activeLimitSortKey},
		"DefaultLimit": &types.AttributeValueMemberN{Value: strconv.Itoa(policy.Default())},
	}
	if overrides := policyOverridesMap(policy); len(overrides) > 0 {
		attrs["Overrides"] = stringMapToAttribute(overrides)
	}

	_, err := s.client.Client().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item:      attrs,
	})
	return err
}

// policyOverridesMap renders each category override as a string for storage.
func policyOverridesMap(policy domain.ActiveLimitPolicy) map[string]string {
	raw := policy.Overrides()
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for category, limit := range raw {
		out[category] = strconv.Itoa(limit)
	}
	return out
}

package dynamodb

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// CreditedRosterReader resolves member_ids for a TEAM placement fact from
// the frozen rosters of exactly the contracts that contributed credits.
// An uncredited contract of the same team and season can never leak its
// members into the awarded fact.
type CreditedRosterReader struct {
	client    *DynamoDBClient
	tableName string
}

// NewCreditedRosterReader creates the adapter implementing
// application.CreditedRosterReader.
func NewCreditedRosterReader(client *DynamoDBClient, tableName string) *CreditedRosterReader {
	return &CreditedRosterReader{client: client, tableName: tableName}
}

// Members returns the deduplicated, sorted union of the frozen rosters of
// the credited contracts, resolved by a strongly consistent read per
// contract id. A missing or unfrozen contract contributes no members.
func (r *CreditedRosterReader) Members(seasonID, category, recipientID string, creditedContractIDs []string) ([]string, error) {
	ctx := context.Background()
	members := map[string]struct{}{}
	for _, id := range creditedContractIDs {
		result, err := r.client.Client().GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(r.tableName),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "CONTRACT"},
				"SK": &types.AttributeValueMemberS{Value: id},
			},
			ConsistentRead: aws.Bool(true),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve credited roster %s: %w", id, err)
		}
		if result.Item == nil {
			continue
		}
		contract, err := mapToContract(result.Item)
		if err != nil {
			return nil, fmt.Errorf("failed to map credited contract %s: %w", id, err)
		}
		if contract.SeasonID != seasonID || !contract.RosterFrozen() {
			continue
		}
		for _, member := range contract.Roster() {
			members[member] = struct{}{}
		}
	}
	out := make([]string, 0, len(members))
	for member := range members {
		out = append(out, member)
	}
	sort.Strings(out)
	return out, nil
}

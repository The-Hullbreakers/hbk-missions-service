package dynamodb

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type snapshotBucketKey struct {
	category string
	track    string
}

type snapshotEntryAcc struct {
	total     int
	contracts map[string]struct{}
}

// AdjudicationSnapshotReader reads the rank-relevant credit rows for a held
// attempt using strongly consistent base-table reads — never the standings
// GSI — so a write that landed before the fence engaged cannot be missed.
type AdjudicationSnapshotReader struct {
	client    *DynamoDBClient
	tableName string
}

// NewAdjudicationSnapshotReader creates the adapter implementing
// application.AdjudicationSnapshotReader.
func NewAdjudicationSnapshotReader(client *DynamoDBClient, tableName string) *AdjudicationSnapshotReader {
	return &AdjudicationSnapshotReader{client: client, tableName: tableName}
}

// Read aggregates MERIT_CREDIT rows of the attempt's season into bucket
// entries: summed totals plus the distinct contract ids that contributed.
// Entries are ordered by total descending, then recipient, so the snapshot
// is deterministic regardless of storage order.
func (r *AdjudicationSnapshotReader) Read(attempt domain.AdjudicationAttempt) (application.AdjudicationSnapshot, error) {
	ctx := context.Background()
	buckets := map[snapshotBucketKey]map[string]*snapshotEntryAcc{}
	var start map[string]types.AttributeValue
	for {
		page, err := r.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(r.tableName),
			KeyConditionExpression: aws.String("PK = :pk"),
			FilterExpression:       aws.String("SeasonID = :season"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: "MERIT_CREDIT"},
				":season": &types.AttributeValueMemberS{Value: attempt.SeasonID},
			},
			ConsistentRead: aws.Bool(true), ExclusiveStartKey: start,
		})
		if err != nil {
			return application.AdjudicationSnapshot{}, fmt.Errorf("failed to read adjudication snapshot: %w", err)
		}
		for _, item := range page.Items {
			key := snapshotBucketKey{category: getString(item, "Category"), track: getString(item, "Track")}
			recipient := getString(item, "RecipientID")
			if key.category == "" || key.track == "" || recipient == "" {
				return application.AdjudicationSnapshot{}, domain.ErrStagedFactSetInvalid
			}
			if buckets[key] == nil {
				buckets[key] = map[string]*snapshotEntryAcc{}
			}
			acc := buckets[key][recipient]
			if acc == nil {
				acc = &snapshotEntryAcc{contracts: map[string]struct{}{}}
				buckets[key][recipient] = acc
			}
			acc.total += getInt(item, "Amount")
			if id := getString(item, "ContractID"); id != "" {
				acc.contracts[id] = struct{}{}
			}
		}
		if len(page.LastEvaluatedKey) == 0 {
			break
		}
		start = page.LastEvaluatedKey
	}
	return snapshotFromAccs(buckets), nil
}

// snapshotFromAccs materializes the aggregation into the provider-neutral
// snapshot with deterministic bucket and entry ordering.
func snapshotFromAccs(buckets map[snapshotBucketKey]map[string]*snapshotEntryAcc) application.AdjudicationSnapshot {
	keys := make([]snapshotBucketKey, 0, len(buckets))
	for k := range buckets {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].category != keys[j].category {
			return keys[i].category < keys[j].category
		}
		return keys[i].track < keys[j].track
	})
	out := application.AdjudicationSnapshot{Buckets: make([]application.AdjudicationBucket, 0, len(keys))}
	for _, k := range keys {
		bucket := application.AdjudicationBucket{Category: k.category, Track: domain.Track(k.track)}
		recipients := make([]string, 0, len(buckets[k]))
		for r := range buckets[k] {
			recipients = append(recipients, r)
		}
		sort.Slice(recipients, func(i, j int) bool {
			a, b := buckets[k][recipients[i]], buckets[k][recipients[j]]
			if a.total != b.total {
				return a.total > b.total
			}
			return recipients[i] < recipients[j]
		})
		for _, r := range recipients {
			acc := buckets[k][r]
			ids := make([]string, 0, len(acc.contracts))
			for id := range acc.contracts {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			bucket.Entries = append(bucket.Entries, application.AdjudicationEntry{
				RecipientID: r, Total: acc.total, CreditedContractIDs: ids,
			})
		}
		out.Buckets = append(out.Buckets, bucket)
	}
	return out
}

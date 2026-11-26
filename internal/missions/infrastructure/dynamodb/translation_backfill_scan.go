package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (b *TranslationBackfiller) inspectPartition(ctx context.Context, partition string, reviewed ReviewedMissionTranslation, apply bool, report *TranslationBackfillReport) error {
	var start map[string]types.AttributeValue
	for {
		page, err := b.client.Client().Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(b.table),
			KeyConditionExpression: aws.String("#pk = :pk"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "PK",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: partition},
			},
			ExclusiveStartKey: start,
			ConsistentRead:    aws.Bool(true),
		})
		if err != nil {
			return fmt.Errorf("failed to inspect %s translation candidates: %w", partition, err)
		}
		for _, item := range page.Items {
			if err := b.inspectItem(ctx, partition, item, reviewed, apply, report); err != nil {
				return err
			}
		}
		if len(page.LastEvaluatedKey) == 0 {
			return nil
		}
		start = page.LastEvaluatedKey
	}
}

func (b *TranslationBackfiller) inspectItem(ctx context.Context, partition string, item map[string]types.AttributeValue, reviewed ReviewedMissionTranslation, apply bool, report *TranslationBackfillReport) error {
	if getString(item, "DefinitionCode") != reviewed.DefinitionCode {
		return nil
	}
	record := BackfillRecord{Partition: partition, ID: getString(item, "SK")}
	if _, alreadyTranslated := item["ContentEn"]; alreadyTranslated {
		return nil
	}
	if reason := backfillCandidate(partition, item, reviewed); reason != "" {
		record.Reason = reason
		report.Unresolved = append(report.Unresolved, record)
		return nil
	}
	report.Eligible = append(report.Eligible, record)
	if !apply {
		return nil
	}
	conflict, err := b.addEnglish(ctx, partition, item, reviewed)
	if err != nil {
		return err
	}
	if conflict {
		record.Reason = "CONCURRENT_CHANGE"
		report.Unresolved = append(report.Unresolved, record)
	} else {
		report.Updated = append(report.Updated, record)
	}
	return nil
}

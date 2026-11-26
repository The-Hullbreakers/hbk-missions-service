package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// addEnglish fences the source snapshot and English absence at write time.
// UpdateItem changes one attribute; a concurrent lifecycle or merit write
// cannot be overwritten by a stale whole-item Put.
func (b *TranslationBackfiller) addEnglish(ctx context.Context, partition string, item map[string]types.AttributeValue, reviewed ReviewedMissionTranslation) (bool, error) {
	english := map[string]types.AttributeValue{}
	putEnglishContent(english, reviewed.English)
	names := map[string]string{"#pk": "PK", "#en": "ContentEn"}
	values := map[string]types.AttributeValue{":en": english["ContentEn"]}
	conditions := []string{"attribute_exists(#pk)", "attribute_not_exists(#en)"}
	fields := []string{"DefinitionCode", "Title", "Description", "Rules"}
	if partition == "DEFINITION" {
		fields = append(fields, "Version", "State", "Objectives")
	} else if partition == "CONTRACT" {
		fields = append(fields, "DefinitionVersion", "Terms", "Phases")
	} else {
		fields = append(fields, "DefinitionVersion", "Objectives")
	}
	for i, field := range fields {
		name, value := fmt.Sprintf("#f%d", i), fmt.Sprintf(":f%d", i)
		names[name] = field
		if stored, exists := item[field]; exists {
			values[value] = stored
			conditions = append(conditions, name+" = "+value)
		} else {
			conditions = append(conditions, "attribute_not_exists("+name+")")
		}
	}
	_, err := b.client.Client().UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(b.table),
		Key: map[string]types.AttributeValue{
			"PK": item["PK"],
			"SK": item["SK"],
		},
		UpdateExpression:          aws.String("SET #en = :en"),
		ConditionExpression:       aws.String(strings.Join(conditions, " AND ")),
		ExpressionAttributeNames:  names,
		ExpressionAttributeValues: values,
	})
	if err != nil {
		var conflict *types.ConditionalCheckFailedException
		if errors.As(err, &conflict) {
			return true, nil
		}
		return false, fmt.Errorf("failed to enrich %s translation: %w", partition, err)
	}
	return false, nil
}

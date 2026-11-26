package dynamodb

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// applyOfferFilter builds the FilterExpression for the pushdown-eligible
// fields (season, categories, modes, availability).
func applyOfferFilter(input *dynamodb.QueryInput, filter domain.OfferListFilter) {
	var clauses []string
	names := map[string]string{}
	values := input.ExpressionAttributeValues

	if filter.SeasonID != "" {
		names["#season"] = "SeasonID"
		values[":season"] = &types.AttributeValueMemberS{Value: filter.SeasonID}
		clauses = append(clauses, "#season = :season")
	}
	if len(filter.Categories) > 0 {
		names["#cat"] = "Category"
		clauses = append(clauses, "#cat IN ("+inValues(values, ":cat", filter.Categories)+")")
	}
	if len(filter.Modes) > 0 {
		names["#mode"] = "Mode"
		raw := make([]string, len(filter.Modes))
		for i, m := range filter.Modes {
			raw[i] = string(m)
		}
		clauses = append(clauses, "#mode IN ("+inValues(values, ":mode", raw)+")")
	}
	if len(clauses) == 0 {
		return
	}
	input.FilterExpression = aws.String(strings.Join(clauses, " AND "))
	input.ExpressionAttributeNames = names
}

func inValues(values map[string]types.AttributeValue, prefix string, items []string) string {
	placeholders := make([]string, len(items))
	for i, v := range items {
		key := fmt.Sprintf("%s%d", prefix, i)
		values[key] = &types.AttributeValueMemberS{Value: v}
		placeholders[i] = key
	}
	return strings.Join(placeholders, ", ")
}

// mapOfferItem decodes a stored offer item without a second read.
func mapOfferItem(item map[string]types.AttributeValue) domain.MissionOffer {
	return domain.MissionOffer{
		OfferID:           getString(item, "SK"),
		Title:             getString(item, "Title"),
		Description:       getString(item, "Description"),
		Category:          getString(item, "Category"),
		Mode:              domain.MissionMode(getString(item, "Mode")),
		CompletionReward:  getInt(item, "CompletionReward"),
		Objectives:        objectivesFromItem(item, "Objectives"),
		Rules:             getStringSlice(item, "Rules"),
		SeasonID:          getString(item, "SeasonID"),
		DefinitionCode:    getString(item, "DefinitionCode"),
		DefinitionVersion: getString(item, "DefinitionVersion"),
		ContentEn:         englishContentFromItem(item),
	}
}

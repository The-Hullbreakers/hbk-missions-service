package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// putEnglishContent writes only authored English snapshots; legacy
// Spanish-only items keep the attribute absent.
func putEnglishContent(item map[string]types.AttributeValue, content domain.LocalizedContent) {
	if content.Title == "" && content.Description == "" && len(content.Rules) == 0 && len(content.Objectives) == 0 {
		return
	}
	objectives := make([]types.AttributeValue, 0, len(content.Objectives))
	for _, objective := range content.Objectives {
		objectives = append(objectives, &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"ObjectiveID": &types.AttributeValueMemberS{Value: objective.ObjectiveID},
			"Name":        &types.AttributeValueMemberS{Value: objective.Name},
			"Description": &types.AttributeValueMemberS{Value: objective.Description},
		}})
	}
	item["ContentEn"] = &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
		"Title":       &types.AttributeValueMemberS{Value: content.Title},
		"Description": &types.AttributeValueMemberS{Value: content.Description},
		"Rules":       stringSliceToAttribute(content.Rules),
		"Objectives":  &types.AttributeValueMemberL{Value: objectives},
	}}
}

// englishContentFromItem leaves historical items with no attribute
// Spanish-only.
func englishContentFromItem(item map[string]types.AttributeValue) domain.LocalizedContent {
	raw, ok := item["ContentEn"].(*types.AttributeValueMemberM)
	if !ok {
		return domain.LocalizedContent{}
	}
	content := domain.LocalizedContent{
		Title:       getString(raw.Value, "Title"),
		Description: getString(raw.Value, "Description"),
		Rules:       getStringSlice(raw.Value, "Rules"),
	}
	objectives, ok := raw.Value["Objectives"].(*types.AttributeValueMemberL)
	if !ok {
		return content
	}
	for _, attribute := range objectives.Value {
		objective, ok := attribute.(*types.AttributeValueMemberM)
		if !ok {
			continue
		}
		content.Objectives = append(content.Objectives, domain.LocalizedObjective{
			ObjectiveID: getString(objective.Value, "ObjectiveID"),
			Name:        getString(objective.Value, "Name"),
			Description: getString(objective.Value, "Description"),
		})
	}
	return content
}

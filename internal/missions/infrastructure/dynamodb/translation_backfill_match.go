package dynamodb

import (
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func backfillCandidate(partition string, item map[string]types.AttributeValue, reviewed ReviewedMissionTranslation) string {
	versionKey, objectiveKey := "DefinitionVersion", "Objectives"
	if partition == "DEFINITION" {
		versionKey = "Version"
	} else if partition == "CONTRACT" {
		objectiveKey = "Terms"
	}
	if getString(item, versionKey) != reviewed.DefinitionVersion {
		return "VERSION_MISMATCH"
	}
	if partition == "DEFINITION" {
		sk := getString(item, "SK")
		if getString(item, "State") != string(domain.DefinitionStatePublished) ||
			(sk != reviewed.DefinitionCode && sk != reviewed.DefinitionCode+"#"+reviewed.DefinitionVersion) {
			return "VERSION_MISMATCH"
		}
	}
	es := reviewed.SourceSpanish
	if getString(item, "Title") != es.Title ||
		getString(item, "Description") != es.Description ||
		!slices.Equal(getStringSlice(item, "Rules"), es.Rules) ||
		!backfillObjectivesMatch(objectivesFromItem(item, objectiveKey), es.Objectives) {
		return "SPANISH_SNAPSHOT_MISMATCH"
	}
	if partition == "CONTRACT" {
		if _, present := item["Phases"]; present {
			phases := phasesFromItem(item, "Phases")
			if len(phases) == 0 {
				return "SPANISH_SNAPSHOT_MISMATCH"
			}
			var phaseObjectives []domain.Objective
			for _, phase := range phases {
				phaseObjectives = append(phaseObjectives, phase.Mandatory...)
				phaseObjectives = append(phaseObjectives, phase.Optional...)
			}
			if !backfillObjectivesMatch(phaseObjectives, es.Objectives) {
				return "SPANISH_SNAPSHOT_MISMATCH"
			}
		}
	}
	return ""
}

func backfillObjectivesMatch(stored []domain.Objective, source []domain.LocalizedObjective) bool {
	if len(stored) != len(source) {
		return false
	}
	byID := make(map[string]domain.LocalizedObjective, len(source))
	for _, objective := range source {
		if _, exists := byID[objective.ObjectiveID]; exists {
			return false
		}
		byID[objective.ObjectiveID] = objective
	}
	seen := make(map[string]bool, len(stored))
	for _, objective := range stored {
		expected, ok := byID[objective.ObjectiveID]
		if !ok || seen[objective.ObjectiveID] ||
			objective.Name != expected.Name || objective.Description != expected.Description {
			return false
		}
		seen[objective.ObjectiveID] = true
	}
	return true
}

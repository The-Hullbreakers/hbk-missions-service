package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// objectivesToAttribute serializes objectives as a DynamoDB list of maps so
// merit, bonus, and unit limit round-trip on definitions and contract terms.
func objectivesToAttribute(objectives []domain.Objective) types.AttributeValue {
	members := make([]types.AttributeValue, 0, len(objectives))
	for _, o := range objectives {
		members = append(members, &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"ObjectiveID":  &types.AttributeValueMemberS{Value: o.ObjectiveID},
			"Name":         &types.AttributeValueMemberS{Value: o.Name},
			"Description":  &types.AttributeValueMemberS{Value: o.Description},
			"Category":     &types.AttributeValueMemberS{Value: o.Category},
			"MeritPerUnit": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", o.MeritPerUnit)},
			"Bonus":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", o.Bonus)},
			"UnitLimit":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", o.UnitLimit)},
		}})
	}
	return &types.AttributeValueMemberL{Value: members}
}

// phasesToAttribute serializes each phase as {mandatory: [...], optional: [...]}
// so mandatory/optional structure round-trips on phased contract terms.
func phasesToAttribute(phases []domain.Phase) types.AttributeValue {
	members := make([]types.AttributeValue, 0, len(phases))
	for _, phase := range phases {
		members = append(members, &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"Mandatory": objectivesToAttribute(phase.Mandatory),
			"Optional":  objectivesToAttribute(phase.Optional),
		}})
	}
	return &types.AttributeValueMemberL{Value: members}
}

// phasesFromItem deserializes the phase list attribute; absent items return nil.
func phasesFromItem(item map[string]types.AttributeValue, key string) []domain.Phase {
	list, ok := item[key].(*types.AttributeValueMemberL)
	if !ok {
		return nil
	}
	phases := make([]domain.Phase, 0, len(list.Value))
	for _, member := range list.Value {
		m, ok := member.(*types.AttributeValueMemberM)
		if !ok {
			continue
		}
		phases = append(phases, domain.Phase{
			Mandatory: objectivesFromItem(m.Value, "Mandatory"),
			Optional:  objectivesFromItem(m.Value, "Optional"),
		})
	}
	return phases
}

// disputesToAttribute serializes each dispute as a map so the opened/resolved
// lifecycle round-trips on the contract item.
func disputesToAttribute(disputes []domain.Dispute) types.AttributeValue {
	members := make([]types.AttributeValue, 0, len(disputes))
	for _, d := range disputes {
		members = append(members, &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"DisputeID":   &types.AttributeValueMemberS{Value: d.DisputeID},
			"ContractID":  &types.AttributeValueMemberS{Value: d.ContractID},
			"ObjectiveID": &types.AttributeValueMemberS{Value: d.ObjectiveID},
			"RaisedBy":    &types.AttributeValueMemberS{Value: d.RaisedBy},
			"Reason":      &types.AttributeValueMemberS{Value: d.Reason},
			"State":       &types.AttributeValueMemberS{Value: string(d.State)},
			"ResolvedBy":  &types.AttributeValueMemberS{Value: d.ResolvedBy},
		}})
	}
	return &types.AttributeValueMemberL{Value: members}
}

// disputesFromItem deserializes the dispute list attribute; absent items
// return nil.
func disputesFromItem(item map[string]types.AttributeValue, key string) []domain.Dispute {
	list, ok := item[key].(*types.AttributeValueMemberL)
	if !ok {
		return nil
	}
	disputes := make([]domain.Dispute, 0, len(list.Value))
	for _, member := range list.Value {
		m, ok := member.(*types.AttributeValueMemberM)
		if !ok {
			continue
		}
		disputes = append(disputes, domain.Dispute{
			DisputeID:   getString(m.Value, "DisputeID"),
			ContractID:  getString(m.Value, "ContractID"),
			ObjectiveID: getString(m.Value, "ObjectiveID"),
			RaisedBy:    getString(m.Value, "RaisedBy"),
			Reason:      getString(m.Value, "Reason"),
			State:       domain.DisputeState(getString(m.Value, "State")),
			ResolvedBy:  getString(m.Value, "ResolvedBy"),
		})
	}
	return disputes
}

// objectivesFromItem deserializes the objective list attribute.
func objectivesFromItem(item map[string]types.AttributeValue, key string) []domain.Objective {
	list, ok := item[key].(*types.AttributeValueMemberL)
	if !ok {
		return nil
	}
	objectives := make([]domain.Objective, 0, len(list.Value))
	for _, member := range list.Value {
		m, ok := member.(*types.AttributeValueMemberM)
		if !ok {
			continue
		}
		objectives = append(objectives, domain.NewObjectiveNamed(
			getString(m.Value, "ObjectiveID"),
			getString(m.Value, "Name"),
			getString(m.Value, "Description"),
			getString(m.Value, "Category"),
			getInt(m.Value, "MeritPerUnit"),
			getInt(m.Value, "Bonus"),
			getInt(m.Value, "UnitLimit"),
		))
	}
	return objectives
}

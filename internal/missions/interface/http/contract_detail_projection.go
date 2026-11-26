package http

import (
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// detailCredit projects a staged credit for reviewer inspection — the
// media_uri is a reference the reviewer opens via the read grant.
type detailCredit struct {
	ObjectiveID string `json:"objective_id"`
	UnitID      string `json:"unit_id"`
	RecipientID string `json:"recipient_id"`
	Amount      int    `json:"amount"`
	MediaURI    string `json:"media_uri,omitempty"`
}

// detailObjective exposes the frozen objective snapshot on the contract.
type detailObjective struct {
	ObjectiveID  string `json:"objective_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	MeritPerUnit int    `json:"merit_per_unit"`
	Bonus        int    `json:"bonus"`
	UnitLimit    int    `json:"unit_limit"`
}

type contractDetail struct {
	ContractID       string            `json:"contract_id"`
	OfferID          string            `json:"offer_id"`
	DefinitionCode   string            `json:"definition_code"`
	Title            string            `json:"title"`
	ContentLanguage  string            `json:"content_language"`
	CompletionReward int               `json:"completion_reward"`
	State            string            `json:"state"`
	Roster           []string          `json:"roster"`
	Departures       []string          `json:"departures"`
	SeasonID         string            `json:"season_id"`
	TeamID           string            `json:"team_id"`
	Terms            []detailObjective `json:"terms"`
	Confirmations    []string          `json:"confirmations"`
	PendingCredits   []detailCredit    `json:"pending_credits"`
	RejectedUnits    []detailUnitRef   `json:"rejected_units"`
	Invitations      map[string]string `json:"invitations,omitempty"`
	Phases           []detailPhase     `json:"phases"`
	CurrentPhase     int               `json:"current_phase"`
}

// detailPhase exposes one contract stage with mandatory/optional objectives.
type detailPhase struct {
	Mandatory []detailObjective `json:"mandatory"`
	Optional  []detailObjective `json:"optional"`
}

func toContractDetail(c *domain.MissionContract) contractDetail {
	terms := objectivesProjection(c.Terms.Objectives())
	phases := make([]detailPhase, 0)
	for _, phase := range c.Terms.Phases() {
		phases = append(phases, detailPhase{
			Mandatory: objectivesProjection(phase.Mandatory),
			Optional:  objectivesProjection(phase.Optional),
		})
	}
	credits := make([]detailCredit, 0)
	for _, credit := range c.PendingCredits() {
		credits = append(credits, detailCredit{
			ObjectiveID: credit.ObjectiveID,
			UnitID:      credit.UnitID,
			RecipientID: credit.RecipientID,
			Amount:      credit.Amount,
			MediaURI:    string(credit.MediaURI),
		})
	}
	return contractDetail{
		ContractID:       c.ContractID,
		OfferID:          c.Offer.OfferID,
		DefinitionCode:   c.Offer.DefinitionCode,
		Title:            c.Offer.Title,
		CompletionReward: c.Offer.CompletionReward,
		State:            string(c.State),
		Roster:           c.Roster(),
		Departures:       c.Departures(),
		SeasonID:         c.SeasonID,
		TeamID:           c.TeamID,
		Terms:            terms,
		Confirmations:    c.Confirmations(),
		PendingCredits:   credits,
		RejectedUnits:    rejectedUnitsProjection(c.RejectedUnits()),
		Invitations:      invitationProjection(c),
		Phases:           phases,
		CurrentPhase:     c.CurrentPhase(),
	}
}

// objectivesProjection maps domain objectives into the detail wire shape.
func objectivesProjection(objectives []domain.Objective) []detailObjective {
	out := make([]detailObjective, 0, len(objectives))
	for _, objective := range objectives {
		out = append(out, detailObjective{
			ObjectiveID:  objective.ObjectiveID,
			Name:         objective.Name,
			Description:  objective.Description,
			Category:     objective.Category,
			MeritPerUnit: objective.MeritPerUnit,
			Bonus:        objective.Bonus,
			UnitLimit:    objective.UnitLimit,
		})
	}
	return out
}

// invitationProjection exposes member invitation states for reviewers; absent
// for solo contracts and pre-invitation data.
func invitationProjection(c *domain.MissionContract) map[string]string {
	states := c.Invitations()
	if len(states) == 0 {
		return nil
	}
	projection := make(map[string]string, len(states))
	for member, state := range states {
		projection[member] = string(state)
	}
	return projection
}

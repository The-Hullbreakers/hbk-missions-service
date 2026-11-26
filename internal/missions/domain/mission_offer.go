package domain

import "encoding/json"

// MissionMode defines how a mission is played and scored.
type MissionMode string

const (
	MissionModeSolo MissionMode = "SOLO"
	MissionModeTeam MissionMode = "TEAM"
	MissionModeFlex MissionMode = "FLEX"
)

// IsValid reports whether the mode is a recognized mission mode.
func (m MissionMode) IsValid() bool {
	return m == MissionModeSolo || m == MissionModeTeam || m == MissionModeFlex
}

// MissionOffer is a published mission that can be accepted. Title,
// Description, CompletionReward, Objectives and Rules freeze the definition
// version's content so the board and contracts render stable mission data.
type MissionOffer struct {
	OfferID           string
	Title             string
	Description       string
	Category          string
	Mode              MissionMode
	CompletionReward  int
	Objectives        []Objective
	Rules             []string
	SeasonID          string
	DefinitionCode    string
	DefinitionVersion string
	ContentEn         LocalizedContent `json:"-"`
}

// MeritSummary reports the maximum merits an accepted contract can earn:
// objective credits plus the frozen completion reward.
func (o MissionOffer) MeritSummary() MeritSummary {
	total := 0
	for _, objective := range o.Objectives {
		total += (objective.MeritPerUnit + objective.Bonus) * objective.UnitLimit
	}
	return MeritSummary{
		ObjectiveMerits:  total,
		CompletionReward: o.CompletionReward,
		TotalMerits:      total + o.CompletionReward,
	}
}

// MarshalJSON renders the offer with its derived merit summary so board
// clients can display the earnable breakdown without a second read.
func (o MissionOffer) MarshalJSON() ([]byte, error) {
	type alias MissionOffer
	return json.Marshal(struct {
		alias
		MeritSummary MeritSummary
	}{alias: alias(o), MeritSummary: o.MeritSummary()})
}

// NewMissionOffer creates a solo offer with the given identity, category and season.
func NewMissionOffer(offerID, category string, seasonID string) MissionOffer {
	return NewMissionOfferWithMode(offerID, category, MissionModeSolo, seasonID)
}

// NewTeamMissionOffer creates a team offer with the given identity, category and season.
func NewTeamMissionOffer(offerID, category string, seasonID string) MissionOffer {
	return NewMissionOfferWithMode(offerID, category, MissionModeTeam, seasonID)
}

// NewMissionOfferWithMode creates an offer with the given mode and season.
func NewMissionOfferWithMode(offerID, category string, mode MissionMode, seasonID string) MissionOffer {
	return MissionOffer{
		OfferID:  offerID,
		Category: category,
		Mode:     mode,
		SeasonID: seasonID,
	}
}

// NewMissionOfferFromDefinition creates an offer from a published definition.
func NewMissionOfferFromDefinition(offerID, seasonID string, definition MissionDefinition) (MissionOffer, error) {
	if offerID == "" || seasonID == "" || definition.DefinitionCode == "" {
		return MissionOffer{}, ErrOfferInvalid
	}
	if !definition.State.IsPublished() {
		return MissionOffer{}, ErrDefinitionNotPublished
	}
	return MissionOffer{
		OfferID:           offerID,
		Title:             definition.Title,
		Description:       definition.Description,
		Category:          definition.Category,
		Mode:              definition.Mode,
		CompletionReward:  definition.CompletionReward,
		Objectives:        append([]Objective(nil), definition.Objectives...),
		Rules:             append([]string(nil), definition.Rules...),
		ContentEn:         definition.ContentEn.Clone(),
		SeasonID:          seasonID,
		DefinitionCode:    definition.DefinitionCode,
		DefinitionVersion: definition.Version,
	}, nil
}

// NewValidMissionOffer creates an offer after validating required fields.
func NewValidMissionOffer(offerID, category string, mode MissionMode, seasonID string) (MissionOffer, error) {
	if offerID == "" || category == "" || seasonID == "" || !mode.IsValid() {
		return MissionOffer{}, ErrOfferInvalid
	}
	return NewMissionOfferWithMode(offerID, category, mode, seasonID), nil
}

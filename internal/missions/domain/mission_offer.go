package domain

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

// MissionOffer is a published mission that can be accepted.
type MissionOffer struct {
	OfferID           string
	Category          string
	Capacity          int
	Mode              MissionMode
	SeasonID          string
	DefinitionID      string
	DefinitionVersion string
}

// NewMissionOffer creates a solo offer with the given identity, category, capacity and season.
func NewMissionOffer(offerID, category string, capacity int, seasonID string) MissionOffer {
	return NewMissionOfferWithMode(offerID, category, capacity, MissionModeSolo, seasonID)
}

// NewTeamMissionOffer creates a team offer with the given identity, category, capacity and season.
func NewTeamMissionOffer(offerID, category string, capacity int, seasonID string) MissionOffer {
	return NewMissionOfferWithMode(offerID, category, capacity, MissionModeTeam, seasonID)
}

// NewMissionOfferWithMode creates an offer with the given mode and season.
func NewMissionOfferWithMode(offerID, category string, capacity int, mode MissionMode, seasonID string) MissionOffer {
	return MissionOffer{
		OfferID:  offerID,
		Category: category,
		Capacity: capacity,
		Mode:     mode,
		SeasonID: seasonID,
	}
}

// NewMissionOfferFromDefinition creates an offer from a published definition.
func NewMissionOfferFromDefinition(offerID, seasonID string, definition MissionDefinition, capacity int) (MissionOffer, error) {
	if offerID == "" || seasonID == "" || definition.DefinitionID == "" {
		return MissionOffer{}, ErrOfferInvalid
	}
	if !definition.State.IsPublished() {
		return MissionOffer{}, ErrDefinitionNotPublished
	}
	if capacity <= 0 {
		capacity = definition.Capacity
	}
	if capacity < 0 {
		return MissionOffer{}, ErrOfferInvalid
	}
	return MissionOffer{
		OfferID:           offerID,
		Category:          definition.Category,
		Capacity:          capacity,
		Mode:              definition.Mode,
		SeasonID:          seasonID,
		DefinitionID:      definition.DefinitionID,
		DefinitionVersion: definition.Version,
	}, nil
}

// NewValidMissionOffer creates an offer after validating required fields.
func NewValidMissionOffer(offerID, category string, capacity int, mode MissionMode, seasonID string) (MissionOffer, error) {
	if offerID == "" || category == "" || seasonID == "" || capacity < 0 || !mode.IsValid() {
		return MissionOffer{}, ErrOfferInvalid
	}
	return NewMissionOfferWithMode(offerID, category, capacity, mode, seasonID), nil
}

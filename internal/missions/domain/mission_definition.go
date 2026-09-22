package domain

// DefinitionState represents the lifecycle state of a mission definition.
type DefinitionState string

const (
	DefinitionStateDraft     DefinitionState = "DRAFT"
	DefinitionStatePublished DefinitionState = "PUBLISHED"
	DefinitionStateRetired   DefinitionState = "RETIRED"
)

// IsValid reports whether the state is a recognized definition state.
func (s DefinitionState) IsValid() bool {
	return s == DefinitionStateDraft || s == DefinitionStatePublished || s == DefinitionStateRetired
}

// IsPublished reports whether the state is PUBLISHED.
func (s DefinitionState) IsPublished() bool {
	return s == DefinitionStatePublished
}

// MissionDefinition is a reusable, versioned mission template.
type MissionDefinition struct {
	DefinitionID string
	Category     string
	Mode         MissionMode
	Capacity     int
	State        DefinitionState
	Version      string
}

// NewMissionDefinition creates a draft definition.
func NewMissionDefinition(definitionID, category string, capacity int, mode MissionMode) (MissionDefinition, error) {
	if definitionID == "" || category == "" || capacity < 0 || !mode.IsValid() {
		return MissionDefinition{}, ErrDefinitionInvalid
	}
	return MissionDefinition{
		DefinitionID: definitionID,
		Category:     category,
		Mode:         mode,
		Capacity:     capacity,
		State:        DefinitionStateDraft,
		Version:      "",
	}, nil
}

// Publish transitions the definition from DRAFT to PUBLISHED with the given version.
func (d *MissionDefinition) Publish(version string) error {
	if d.State != DefinitionStateDraft {
		return ErrDefinitionInvalid
	}
	if version == "" {
		return ErrDefinitionInvalid
	}
	d.State = DefinitionStatePublished
	d.Version = version
	return nil
}

// Retire transitions the definition from PUBLISHED to RETIRED.
func (d *MissionDefinition) Retire() error {
	if d.State != DefinitionStatePublished {
		return ErrDefinitionInvalid
	}
	d.State = DefinitionStateRetired
	return nil
}

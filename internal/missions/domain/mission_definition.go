package domain

import "strings"

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

// MissionDefinition is a reusable, versioned mission template with
// player-facing content: title, markdown description, named objectives,
// readable rules, and the competitive-points reward granted when a
// contract on it completes.
type MissionDefinition struct {
	DefinitionCode   string
	Title            string
	Description      string
	Category         string
	Mode             MissionMode
	CompletionReward int
	State            DefinitionState
	Version          string
	Objectives       []Objective
	Rules            []string
	ContentEn        LocalizedContent `json:"-"`
}

// NewMissionDefinition creates a draft definition with required player-facing
// content. Objectives are validated element-wise and require a name; rules
// are optional short requirement statements and reject blank items.
func NewMissionDefinition(definitionCode, title, description, category string, mode MissionMode, completionReward int, objectives []Objective, rules []string) (MissionDefinition, error) {
	if definitionCode == "" || title == "" || description == "" || category == "" ||
		!mode.IsValid() || completionReward < 0 {
		return MissionDefinition{}, ErrDefinitionInvalid
	}
	for _, objective := range objectives {
		if objective.ObjectiveID == "" || objective.Name == "" ||
			objective.ObjectiveID == CompletionObjectiveID {
			return MissionDefinition{}, ErrDefinitionInvalid
		}
	}
	for _, rule := range rules {
		if strings.TrimSpace(rule) == "" {
			return MissionDefinition{}, ErrDefinitionInvalid
		}
	}
	return MissionDefinition{
		DefinitionCode:   definitionCode,
		Title:            title,
		Description:      description,
		Category:         category,
		Mode:             mode,
		CompletionReward: completionReward,
		State:            DefinitionStateDraft,
		Version:          "",
		Objectives:       objectives,
		Rules:            append([]string(nil), rules...),
	}, nil
}

// Publish transitions the definition from DRAFT to PUBLISHED with the given
// version. A new version publishes only with a structurally complete
// English snapshot; already-published records are never revalidated.
func (d *MissionDefinition) Publish(version string) error {
	if d.State != DefinitionStateDraft {
		return ErrDefinitionInvalid
	}
	if version == "" {
		return ErrDefinitionInvalid
	}
	if !d.englishSnapshotComplete() {
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

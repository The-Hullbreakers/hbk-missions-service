package domain

// LocalizedContent holds the player-facing text of one locale. Spanish
// stays on the parent record's scalar fields; other locales are optional
// translation snapshots frozen alongside them.
type LocalizedContent struct {
	Title       string
	Description string
	Rules       []string
	Objectives  []LocalizedObjective
}

// LocalizedObjective is the translated name and description of one
// objective, keyed by its stable ObjectiveID. Economics remain on
// Objective; this record carries player-facing text only.
type LocalizedObjective struct {
	ObjectiveID string
	Name        string
	Description string
}

// Clone returns a deep copy so a frozen snapshot never aliases the
// source's slices. Nil slices remain nil.
func (c LocalizedContent) Clone() LocalizedContent {
	c.Rules = append([]string(nil), c.Rules...)
	c.Objectives = append([]LocalizedObjective(nil), c.Objectives...)
	return c
}

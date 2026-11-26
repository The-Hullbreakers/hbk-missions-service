package domain

// Objective is a published unit of work that grants merits when confirmed.
type Objective struct {
	ObjectiveID  string
	Name         string
	Description  string
	Category     string
	MeritPerUnit int
	Bonus        int
	UnitLimit    int
}

// NewObjectiveNamed creates an objective with full content and scoring fields.
func NewObjectiveNamed(id, name, description, category string, meritPerUnit, bonus, unitLimit int) Objective {
	return Objective{
		ObjectiveID:  id,
		Name:         name,
		Description:  description,
		Category:     category,
		MeritPerUnit: meritPerUnit,
		Bonus:        bonus,
		UnitLimit:    unitLimit,
	}
}

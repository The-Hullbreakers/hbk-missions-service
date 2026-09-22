package domain

// Objective is a published unit of work that grants merits when confirmed.
type Objective struct {
	ObjectiveID  string
	Category     string
	MeritPerUnit int
	Bonus        int
	UnitLimit    int
}

// NewObjective creates an objective with the given identity, category and merit per unit.
func NewObjective(id, category string, meritPerUnit int) Objective {
	return NewObjectiveWithBonusAndLimit(id, category, meritPerUnit, 0, 0)
}

// NewObjectiveWithBonus creates an objective with an optional per-unit bonus.
func NewObjectiveWithBonus(id, category string, meritPerUnit, bonus int) Objective {
	return NewObjectiveWithBonusAndLimit(id, category, meritPerUnit, bonus, 0)
}

// NewObjectiveWithBonusAndLimit creates an objective with an optional per-unit bonus and a unit limit.
func NewObjectiveWithBonusAndLimit(id, category string, meritPerUnit, bonus, unitLimit int) Objective {
	return Objective{
		ObjectiveID:  id,
		Category:     category,
		MeritPerUnit: meritPerUnit,
		Bonus:        bonus,
		UnitLimit:    unitLimit,
	}
}

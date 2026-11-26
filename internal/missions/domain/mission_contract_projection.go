package domain

import "encoding/json"

// MarshalJSON renders the contract with its embedded offer projected from
// the frozen terms when the offer itself carries no objectives. Persisted
// contract items store objectives only under Terms, so without this
// hydration offer.objectives and the derived offer.merit_summary would
// serialize empty even though the contract promises merits (REG-0047).
func (c MissionContract) MarshalJSON() ([]byte, error) {
	type alias MissionContract
	projected := c
	if len(projected.Offer.Objectives) == 0 {
		projected.Offer.Objectives = c.Terms.Objectives()
	}
	return json.Marshal(alias(projected))
}

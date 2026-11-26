package domain

// Roster returns a defensive copy of the frozen roster.
func (c *MissionContract) Roster() []string {
	copied := make([]string, len(c.roster))
	copy(copied, c.roster)
	return copied
}

// IncludesPlayer reports whether the player is part of the contract roster.
func (c *MissionContract) IncludesPlayer(playerID string) bool {
	for _, p := range c.roster {
		if p == playerID {
			return true
		}
	}
	return false
}

// IsDeparted reports whether the player left this contract.
func (c *MissionContract) IsDeparted(playerID string) bool {
	return c.departed[playerID]
}

// ActiveMembers returns roster members who have not departed.
func (c *MissionContract) ActiveMembers() []string {
	active := make([]string, 0, len(c.roster))
	for _, p := range c.roster {
		if !c.departed[p] {
			active = append(active, p)
		}
	}
	return active
}

// Departures returns a defensive copy of the departed member ids.
func (c *MissionContract) Departures() []string {
	departed := make([]string, 0, len(c.departed))
	for id := range c.departed {
		departed = append(departed, id)
	}
	return departed
}

// RosterFrozen returns true once the roster has been frozen at the start of the contract.
func (c *MissionContract) RosterFrozen() bool {
	return c.rosterFrozen
}

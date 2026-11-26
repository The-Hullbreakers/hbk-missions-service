package domain

// Roster is the immutable set of players committed to a contract: the
// requester plus the validated teammates. Members are non-empty and distinct.
type Roster struct {
	members []string
}

// NewRoster builds a roster from the requester and their teammates.
// The requester is always the first member; teammates may not repeat
// the requester or each other.
func NewRoster(requester string, teammates []string) (Roster, error) {
	if requester == "" {
		return Roster{}, ErrRosterEmpty
	}
	members := make([]string, 0, len(teammates)+1)
	seen := make(map[string]bool, len(teammates)+1)
	members = append(members, requester)
	seen[requester] = true
	for _, teammate := range teammates {
		if teammate == "" {
			return Roster{}, ErrRosterEmpty
		}
		if seen[teammate] {
			return Roster{}, ErrRosterDuplicate
		}
		seen[teammate] = true
		members = append(members, teammate)
	}
	return Roster{members: members}, nil
}

// Members returns a defensive copy of the roster, requester first.
func (r Roster) Members() []string {
	return append([]string(nil), r.members...)
}

// Contains reports whether the player is a roster member.
func (r Roster) Contains(playerID string) bool {
	for _, member := range r.members {
		if member == playerID {
			return true
		}
	}
	return false
}

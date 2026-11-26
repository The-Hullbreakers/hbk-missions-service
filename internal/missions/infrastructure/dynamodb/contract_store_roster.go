package dynamodb

import (
	"sort"
)

// Members returns the deduplicated union of the frozen rosters of the team's
// contracts in the season — the snapshot captured when each contract
// started. Teams without a frozen roster in the season yield an empty list.
func (s *ContractStore) Members(seasonID, teamID string) ([]string, error) {
	contracts, err := s.listAll()
	if err != nil {
		return nil, err
	}

	members := map[string]bool{}
	for _, contract := range contracts {
		if contract.TeamID != teamID || contract.SeasonID != seasonID || !contract.RosterFrozen() {
			continue
		}
		for _, member := range contract.Roster() {
			members[member] = true
		}
	}

	out := make([]string, 0, len(members))
	for member := range members {
		out = append(out, member)
	}
	sort.Strings(out)
	return out, nil
}

package domain

import (
	"sort"
	"strings"
)

// MembersOfCreditedContracts resolves the personal grant recipients for a
// TEAM placement fact. Only the frozen rosters of contracts whose credits
// actually fed that exact bucket/recipient contribute members — an
// uncredited contract's roster can never leak in. Every member on a
// credited roster receives a personal grant even when not individually
// credited. The result is a deduplicated, sorted fresh slice.
func MembersOfCreditedContracts(creditedIDs []string, rosters map[string][]string) ([]string, error) {
	if len(creditedIDs) == 0 {
		return nil, ErrCreditedRosterInvalid
	}
	seen := make(map[string]struct{})
	for _, id := range creditedIDs {
		if strings.TrimSpace(id) == "" {
			return nil, ErrCreditedRosterInvalid
		}
		roster, ok := rosters[id]
		if !ok || len(roster) == 0 {
			return nil, ErrCreditedRosterInvalid
		}
		for _, member := range roster {
			if strings.TrimSpace(member) == "" {
				return nil, ErrCreditedRosterInvalid
			}
			seen[member] = struct{}{}
		}
	}
	members := make([]string, 0, len(seen))
	for member := range seen {
		members = append(members, member)
	}
	sort.Strings(members)
	return members, nil
}

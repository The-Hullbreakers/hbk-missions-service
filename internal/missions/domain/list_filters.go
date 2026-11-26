package domain

import "strings"

// ContractStatusFilter is a business bucket for listing player contracts.
// ALL and ACTIVE are buckets; any ContractState value is an exact-state filter.
type ContractStatusFilter string

const (
	// ContractStatusAll lists contracts in every state (the default).
	ContractStatusAll ContractStatusFilter = "ALL"
	// ContractStatusActive lists non-terminal contracts.
	ContractStatusActive ContractStatusFilter = "ACTIVE"
)

// ContractListQuery carries the raw HTTP query parameters for listing contracts.
type ContractListQuery struct {
	SeasonID string
	Status   string
}

// ContractListFilter is the parsed, typed query for a player's contracts.
type ContractListFilter struct {
	SeasonID string
	Status   ContractStatusFilter
}

// ParseContractListFilter validates the raw query into a typed filter.
// Empty status defaults to ALL. Unknown values return an error.
func ParseContractListFilter(raw ContractListQuery) (ContractListFilter, error) {
	filter := ContractListFilter{SeasonID: raw.SeasonID, Status: ContractStatusAll}
	status := strings.ToUpper(strings.TrimSpace(raw.Status))
	switch ContractStatusFilter(status) {
	case "", ContractStatusAll, ContractStatusActive:
		if status != "" {
			filter.Status = ContractStatusFilter(status)
		}
	case ContractStatusFilter(ContractAccepted),
		ContractStatusFilter(ContractInProgress),
		ContractStatusFilter(ContractPendingValidation),
		ContractStatusFilter(ContractCompleted),
		ContractStatusFilter(ContractFailed),
		ContractStatusFilter(ContractAbandoned),
		ContractStatusFilter(ContractExpired),
		ContractStatusFilter(ContractCancelled),
		ContractStatusFilter(ContractDisqualified):
		filter.Status = ContractStatusFilter(status)
	default:
		return ContractListFilter{}, ErrInvalidContractStatusFilter
	}
	return filter, nil
}

// Match reports whether the contract satisfies the filter.
func (f ContractListFilter) Match(contract MissionContract) bool {
	if f.SeasonID != "" && contract.SeasonID != f.SeasonID {
		return false
	}
	switch f.Status {
	case "", ContractStatusAll:
		return true
	case ContractStatusActive:
		return contract.State.BlocksRetake()
	default:
		return contract.State == ContractState(f.Status)
	}
}

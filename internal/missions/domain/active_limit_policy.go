package domain

// ActiveLimitPolicy defines how many active contracts a player may hold.
// The default applies unless a category override exists.
type ActiveLimitPolicy struct {
	defaultLimit int
	overrides    map[string]int
}

// NewActiveLimitPolicy creates a policy with the given default and optional per-category overrides.
func NewActiveLimitPolicy(defaultLimit int, overrides map[string]int) ActiveLimitPolicy {
	if overrides == nil {
		overrides = map[string]int{}
	}
	return ActiveLimitPolicy{
		defaultLimit: defaultLimit,
		overrides:    overrides,
	}
}

// Default returns the default active contract limit.
func (p ActiveLimitPolicy) Default() int {
	return p.defaultLimit
}

// EffectiveLimit returns the limit for a given category, using an override when present.
func (p ActiveLimitPolicy) EffectiveLimit(category string) int {
	if limit, ok := p.overrides[category]; ok {
		return limit
	}
	return p.defaultLimit
}

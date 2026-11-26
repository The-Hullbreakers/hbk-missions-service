package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// ActiveLimitPolicyWriter persists the active-contract limit document.
type ActiveLimitPolicyWriter interface {
	SaveActiveLimit(policy domain.ActiveLimitPolicy) error
}

// ActiveLimitPolicyReader loads the active-contract limit document.
type ActiveLimitPolicyReader interface {
	GetActiveLimit() (domain.ActiveLimitPolicy, error)
}

// UpdateActiveLimitHandler persists a runtime-tuned active-contract limit;
// non-positive values surface ErrInvalidPolicy.
type UpdateActiveLimitHandler struct {
	writer ActiveLimitPolicyWriter
}

// NewUpdateActiveLimitHandler creates a handler bound to the policy writer.
func NewUpdateActiveLimitHandler(writer ActiveLimitPolicyWriter) *UpdateActiveLimitHandler {
	return &UpdateActiveLimitHandler{writer: writer}
}

// Handle validates and persists the default limit plus optional category
// overrides.
func (h *UpdateActiveLimitHandler) Handle(defaultLimit int, overrides map[string]int) error {
	policy := domain.NewActiveLimitPolicy(defaultLimit, overrides)
	if policy.Default() < 1 {
		return domain.ErrInvalidPolicy
	}
	for _, limit := range policy.Overrides() {
		if limit < 1 {
			return domain.ErrInvalidPolicy
		}
	}
	return h.writer.SaveActiveLimit(policy)
}

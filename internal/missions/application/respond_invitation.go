package application

import (
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// ConfirmInvitationHandler orchestrates the invitation-confirmation use case.
type ConfirmInvitationHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewConfirmInvitationHandler creates a handler with the required ports.
func NewConfirmInvitationHandler(contracts ContractReader, writer ContractWriter) *ConfirmInvitationHandler {
	return &ConfirmInvitationHandler{contracts: contracts, writer: writer}
}

// Handle resolves the caller's pending invitation as confirmed and persists it.
func (h *ConfirmInvitationHandler) Handle(identity domain.CallerIdentity, contractID string) error {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	if err := contract.ConfirmInvitation(identity, time.Now()); err != nil {
		return err
	}
	return h.writer.Save(contract)
}

// DeclineInvitationHandler orchestrates the invitation-decline use case.
type DeclineInvitationHandler struct {
	contracts ContractReader
	writer    ContractWriter
}

// NewDeclineInvitationHandler creates a handler with the required ports.
func NewDeclineInvitationHandler(contracts ContractReader, writer ContractWriter) *DeclineInvitationHandler {
	return &DeclineInvitationHandler{contracts: contracts, writer: writer}
}

// Handle resolves the caller's pending invitation as declined, releasing the
// reserved roster slot, and persists the contract.
func (h *DeclineInvitationHandler) Handle(identity domain.CallerIdentity, contractID string) error {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	if err := contract.DeclineInvitation(identity, time.Now()); err != nil {
		return err
	}
	return h.writer.Save(contract)
}

package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// DepartureWriter persists a member departure, releasing the departing
// member's guards atomically with the contract write. previousState is
// the contract state the departure was decided from; the write must fail
// if the persisted state has since moved past it.
type DepartureWriter interface {
	SaveDeparture(contract *domain.MissionContract, departingPlayerID string, previousState domain.ContractState) error
}

// LeaveContractHandler orchestrates the member-leave use case.
type LeaveContractHandler struct {
	contracts ContractReader
	departure DepartureWriter
}

// NewLeaveContractHandler creates a handler with the required ports.
func NewLeaveContractHandler(contracts ContractReader, departure DepartureWriter) *LeaveContractHandler {
	return &LeaveContractHandler{contracts: contracts, departure: departure}
}

// Handle removes the caller from the contract's active membership on a
// non-terminal contract, persisting the departure; the last active member
// leaving abandons the contract through the same focused write so its
// guards are released too.
func (h *LeaveContractHandler) Handle(identity domain.CallerIdentity, contractID string) error {
	contract, err := h.contracts.Get(contractID)
	if err != nil {
		return err
	}
	previousState := contract.State
	if err := contract.Leave(identity); err != nil {
		return err
	}
	return h.departure.SaveDeparture(contract, identity.PlayerID(), previousState)
}

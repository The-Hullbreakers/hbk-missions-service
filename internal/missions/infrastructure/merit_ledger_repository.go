package infrastructure

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// MeritLedgerRepository is an in-memory adapter for the application MeritLedgerWriter port.
type MeritLedgerRepository struct {
	ledger *domain.MeritLedger
}

// NewMeritLedgerRepository creates a repository backed by the given ledger.
func NewMeritLedgerRepository(ledger *domain.MeritLedger) *MeritLedgerRepository {
	return &MeritLedgerRepository{ledger: ledger}
}

// Save records a credit in the ledger.
func (r *MeritLedgerRepository) Save(credit domain.MeritCredit) error {
	r.ledger.Record(credit)
	return nil
}

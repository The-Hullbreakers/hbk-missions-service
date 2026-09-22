package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestConfirmationID_New(t *testing.T) {
	id, err := domain.NewConfirmationID("contract-1", "objective-1", "unit-1", "player-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id.String() != "contract-1:objective-1:unit-1:player-1" {
		t.Fatalf("unexpected string form: %s", id.String())
	}
	if id.ContractID() != "contract-1" || id.ObjectiveID() != "objective-1" || id.UnitID() != "unit-1" || id.RecipientID() != "player-1" {
		t.Fatalf("unexpected parts: %+v", id)
	}
}

func TestConfirmationID_NewRejectsEmptyPart(t *testing.T) {
	_, err := domain.NewConfirmationID("", "objective-1", "unit-1", "player-1")
	if !errors.Is(err, domain.ErrInvalidConfirmationID) {
		t.Fatalf("expected ErrInvalidConfirmationID, got %v", err)
	}
}

func TestConfirmationID_Parse(t *testing.T) {
	id, err := domain.ParseConfirmationID("contract-1:objective-1:unit-1:player-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id.ContractID() != "contract-1" {
		t.Fatalf("unexpected contract id: %s", id.ContractID())
	}
}

func TestConfirmationID_ParseRejectsMalformed(t *testing.T) {
	_, err := domain.ParseConfirmationID("c:1")
	if !errors.Is(err, domain.ErrInvalidConfirmationID) {
		t.Fatalf("expected ErrInvalidConfirmationID, got %v", err)
	}
}

func TestConfirmationID_IsZero(t *testing.T) {
	var zero domain.ConfirmationID
	if !zero.IsZero() {
		t.Fatalf("expected zero value to be zero")
	}
	id, _ := domain.NewConfirmationID("c", "o", "u", "r")
	if id.IsZero() {
		t.Fatalf("expected non-zero value")
	}
}

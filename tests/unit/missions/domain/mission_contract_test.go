package domain_test

import (
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestMissionContract_Summary(t *testing.T) {
	offer := domain.MissionOffer{OfferID: "offer-1", Category: "BOUNTY_HUNTING", SeasonID: "season-1"}
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})

	summary := contract.Summary()

	if summary.OfferID != "offer-1" || summary.Category != "BOUNTY_HUNTING" || summary.State != domain.ContractAccepted {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestMissionContract_StartFreezesTeamRoster(t *testing.T) {
	// Arrange
	offer := domain.MissionOffer{OfferID: "offer-1", Category: "BOUNTY_HUNTING", SeasonID: "season-1"}
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1", "player-2"})

	// Act
	startContract(t, contract)

	// Assert
	if contract.State != domain.ContractInProgress {
		t.Fatalf("expected state IN_PROGRESS, got %s", contract.State)
	}
	if !contract.RosterFrozen() {
		t.Fatalf("expected roster to be frozen after Start")
	}
}

func TestMissionContract_FromStatePreservesConfirmations(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContractFromState("contract-1", offer, []string{"player-1"}, domain.ContractInProgress, true, "", "season-1", []string{"c:1"}, nil, nil)

	if len(contract.Confirmations()) != 1 || contract.Confirmations()[0] != "c:1" {
		t.Fatalf("expected confirmation c:1, got %v", contract.Confirmations())
	}
}

func TestMissionContract_RosterIsDefensivelyCopied(t *testing.T) {
	// Arrange
	offer := domain.MissionOffer{OfferID: "offer-1", Category: "BOUNTY_HUNTING", SeasonID: "season-1"}
	roster := []string{"player-1", "player-2"}
	contract := domain.NewMissionContract("contract-1", offer, roster)

	// Act
	roster[0] = "tampered"

	// Assert
	got := contract.Roster()
	if len(got) != 2 || got[0] != "player-1" {
		t.Fatalf("expected roster unchanged [player-1 player-2], got %v", got)
	}
}

func TestMissionContract_RosterCopyDoesNotMutateInternalRoster(t *testing.T) {
	// Arrange
	offer := domain.MissionOffer{OfferID: "offer-1", Category: "BOUNTY_HUNTING", SeasonID: "season-1"}
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1", "player-2"})

	// Act
	returned := contract.Roster()
	returned[0] = "tampered"

	// Assert
	got := contract.Roster()
	if len(got) != 2 || got[0] != "player-1" {
		t.Fatalf("expected internal roster unchanged [player-1 player-2], got %v", got)
	}
}

func TestMissionContract_AbandonMakesContractTerminal(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	startContract(t, contract)

	// Act
	contract.Abandon(player(t, "player-1"))

	// Assert
	if contract.State != domain.ContractAbandoned {
		t.Fatalf("expected state ABANDONED, got %s", contract.State)
	}
	if !contract.State.IsTerminal() {
		t.Fatalf("expected ABANDONED to be terminal")
	}
}

func TestConfirmObjective_TeamContractCreditsTeamTrack(t *testing.T) {
	// Arrange
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewTeamMissionContract("contract-1", offer, []string{"player-1", "player-2"}, "team-1")
	startContract(t, contract)
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	freezeTerms(t, contract, objective)

	// Act
	credit, err := domain.ConfirmObjective(contract, objective.ObjectiveID, "target-1", season)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if credit.Track != domain.TrackTeam {
		t.Fatalf("expected track TEAM, got %s", credit.Track)
	}
	if credit.RecipientID != "team-1" {
		t.Fatalf("expected recipient team-1, got %s", credit.RecipientID)
	}
}

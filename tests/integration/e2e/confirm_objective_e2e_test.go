package e2e_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type fixedContractStore struct{ contract *domain.MissionContract }

func (s fixedContractStore) Get(string) (*domain.MissionContract, error) { return s.contract, nil }

type fixedContractWriter struct{}

func (fixedContractWriter) Save(*domain.MissionContract) error { return nil }

type fixedSeasonStore struct{ season domain.Season }

func (s fixedSeasonStore) Get(string) (domain.Season, error) { return s.season, nil }

type fixedObjectiveStore struct{ objective domain.Objective }

func (s fixedObjectiveStore) Get(string) (domain.Objective, error) { return s.objective, nil }

func TestConfirmObjectiveE2E_CreditsMerit(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", 10, "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	contract.Start()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	objective := domain.NewObjective("objective-1", "BOUNTY_HUNTING", 100)
	ledger := domain.NewMeritLedger()
	repo := infrastructure.NewMeritLedgerRepository(ledger)

	app := application.NewConfirmObjectiveHandler(
		fixedContractStore{contract: contract},
		fixedContractWriter{},
		fixedSeasonStore{season: season},
		fixedObjectiveStore{objective: objective},
		repo,
	)

	server := httptest.NewServer(missionhttp.ConfirmEvidenceHandler(app))
	defer server.Close()

	body := `{"objective_id":"objective-1","unit_id":"target-1","season_id":"season-1","tenant_id":"tenant-1"}`
	resp, err := http.Post(server.URL+"/api/v1/missions/contracts/contract-1/evidence", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	if ledger.TotalFor("tenant-1", "season-1", "BOUNTY_HUNTING", domain.TrackIndividual, "player-1") != 100 {
		t.Fatalf("expected ledger total 100")
	}
}

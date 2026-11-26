package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
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

func newTestServer(started bool) (*httptest.Server, *stagedContractWriter) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	terms, _ := domain.NewContractTerms([]domain.Objective{objective})
	contract.Terms = terms
	if started {
		member, _ := domain.NewCallerIdentity("player-1", nil, nil)

		contract.Start(member)
	}
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	writer := &stagedContractWriter{}

	app := application.NewConfirmObjectiveHandler(
		fixedContractStore{contract: contract},
		writer,
		fixedSeasonStore{season: season},
	)

	auth := missionhttp.NewAuthMiddleware(roleSessionFake{playerID: "player-1"})
	return httptest.NewServer(auth.Authenticate(missionhttp.ConfirmEvidenceHandler(app))), writer
}

func postEvidence(t *testing.T, server *httptest.Server, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/missions/contracts/contract-1/evidence", strings.NewReader(body))
	if err != nil {
		t.Fatalf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	return resp
}

func TestConfirmEvidenceEndpoint_AcceptsValidEvidence(t *testing.T) {
	server, writer := newTestServer(true)
	defer server.Close()

	body := `{"objective_id":"objective-1","unit_id":"target-1","season_id":"season-1"}`
	resp := postEvidence(t, server, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	if writer.saved == nil || len(writer.saved.PendingCredits()) != 1 || writer.saved.PendingCredits()[0].Amount != 100 {
		t.Fatalf("expected one staged credit of 100")
	}
}

func TestConfirmEvidenceEndpoint_RejectsInvalidJSON(t *testing.T) {
	server, _ := newTestServer(true)
	defer server.Close()

	resp := postEvidence(t, server, "not json")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestConfirmEvidenceEndpoint_RejectsDomainError(t *testing.T) {
	server, _ := newTestServer(false)
	defer server.Close()

	body := `{"objective_id":"objective-1","unit_id":"target-1","season_id":"season-1","tenant_id":"tenant-1"}`
	resp := postEvidence(t, server, body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", resp.StatusCode)
	}

	var bodyJSON map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&bodyJSON); err != nil {
		t.Fatalf("could not decode error body: %v", err)
	}
	if bodyJSON["code"] != "CONTRACT_NOT_IN_PROGRESS" {
		t.Fatalf("expected code CONTRACT_NOT_IN_PROGRESS, got %s", bodyJSON["code"])
	}
}

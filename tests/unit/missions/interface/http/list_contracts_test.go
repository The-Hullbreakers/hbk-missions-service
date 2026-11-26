package http_test

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type memberSession struct{}

func (s memberSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", map[string]any{}, nil
}

func (s memberSession) HasRole(map[string]any, string) bool { return true }

type fakePlayerContractsReaderForHTTP struct {
	contracts []domain.MissionContract
	gotPlayer string
	gotFilter domain.ContractListFilter
}

func (r *fakePlayerContractsReaderForHTTP) ListByPlayer(playerID string, filter domain.ContractListFilter) ([]domain.MissionContract, error) {
	r.gotPlayer = playerID
	r.gotFilter = filter
	return r.contracts, nil
}

func TestListContractsHandler_ReturnsPlayerContracts(t *testing.T) {
	// Arrange
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	mine := domain.NewMissionContract("c1", offer, []string{"player-1"})
	reader := &fakePlayerContractsReaderForHTTP{contracts: []domain.MissionContract{*mine}}
	app := application.NewListPlayerContractsHandler(reader)
	auth := missionhttp.NewAuthMiddleware(memberSession{})
	handler := auth.Authenticate(missionhttp.ListContractsHandler(app))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/contracts", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "player-1", reader.gotPlayer)
	var result []map[string]any
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	require.False(t, len(result) != 1 || result[0]["ContractID"] != "c1")
}

type failingContractsReader struct{}

func (r *failingContractsReader) ListByPlayer(string, domain.ContractListFilter) ([]domain.MissionContract, error) {
	return nil, errors.New("store unavailable")
}

func TestListContractsHandler_Returns500OnReaderError(t *testing.T) {
	// Arrange
	app := application.NewListPlayerContractsHandler(&failingContractsReader{})
	handler := missionhttp.ListContractsHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/contracts", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestListContractsHandler_RejectsUnknownStatus(t *testing.T) {
	// Arrange
	reader := &fakePlayerContractsReaderForHTTP{}
	app := application.NewListPlayerContractsHandler(reader)
	handler := missionhttp.ListContractsHandler(app)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/missions/contracts?status=LOST", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

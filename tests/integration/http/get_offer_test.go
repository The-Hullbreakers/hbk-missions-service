package http_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// offerReaderDouble is an OfferReader stub returning a fixture or error.
type offerReaderDouble struct {
	offer domain.MissionOffer
	err   error
}

func (r *offerReaderDouble) Get(string) (domain.MissionOffer, error) {
	return r.offer, r.err
}

// offerServer mirrors the composition wiring for the single-offer route.
func offerServer(session boundarySession, reader *offerReaderDouble) *httptest.Server {
	app := application.NewGetOfferHandler(reader)
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("GET /api/v1/missions/offers/{offer_id}",
		auth.Authenticate(missionhttp.GetOfferHandler(app)))
	return httptest.NewServer(mux)
}

func getOffer(t *testing.T, server *httptest.Server, offerID string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet,
		server.URL+"/api/v1/missions/offers/"+offerID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// A member fetches a published offer by id.
func TestGetOffer_ReturnsOffer(t *testing.T) {
	// Arrange
	reader := &offerReaderDouble{offer: domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")}
	server := offerServer(boundarySession{playerID: "player-1"}, reader)
	defer server.Close()

	// Act
	resp := getOffer(t, server, "offer-1")

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.False(t, body["OfferID"] != "offer-1" && body["offer_id"] != "offer-1")
}

// An unknown offer id returns 404 OFFER_NOT_AVAILABLE.
func TestGetOffer_UnknownReturns404(t *testing.T) {
	// Arrange
	reader := &offerReaderDouble{err: domain.ErrOfferNotAvailable}
	server := offerServer(boundarySession{playerID: "player-1"}, reader)
	defer server.Close()

	// Act
	resp := getOffer(t, server, "offer-x")

	// Assert
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "OFFER_NOT_AVAILABLE", body["code"])
}

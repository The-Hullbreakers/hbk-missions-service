package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// Fail-closed contract: a request carrying no caller identity is rejected
// with 401 UNAUTHORIZED before the application handler is invoked. The nil
// application double proves the boundary never dereferences it.
func TestAcceptMissionHandler_MissingCallerIdentityReturns401(t *testing.T) {
	// Arrange
	handler := missionhttp.AcceptMissionHandler(nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/missions/offers/offer-1/accept", nil)

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	require.Equal(t, http.StatusUnauthorized, rec.Code)
	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, "UNAUTHORIZED", body["code"])
}

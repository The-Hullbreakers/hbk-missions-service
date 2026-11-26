package http_test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/session"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// stringClaimsSession returns roles/team_ids as single strings rather than
// arrays — a claim shape real OIDC providers emit.
type stringClaimsSession struct{}

var _ session.UserSession = stringClaimsSession{}

func (stringClaimsSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", map[string]any{
		"roles":    "admin",
		"team_ids": "team-9",
	}, nil
}

func (stringClaimsSession) HasRole(map[string]any, string) bool { return true }

func TestCallerIdentityFromContext_SingleStringClaims(t *testing.T) {
	var identity domain.CallerIdentity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		identity, err = missionhttp.CallerIdentityFromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := missionhttp.NewAuthMiddleware(stringClaimsSession{})
	server := httptest.NewServer(middleware.Authenticate(next))
	defer server.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/any", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer any")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, !identity.HasRole("admin"))
	require.False(t, !identity.IsTeamMember("team-9"))
}

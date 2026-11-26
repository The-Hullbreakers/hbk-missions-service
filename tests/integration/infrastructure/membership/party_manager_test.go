package membership_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	"github.com/hbk/hbk-missions-service/internal/missions/infrastructure/membership"
	"github.com/stretchr/testify/require"
)

// PartyManagerReader contract: Party Manager answers GET
// /teams/{team_id}/members/{player_id}; 200 is a member, 404 is a definitive
// non-member, and anything else fails closed as ErrMembershipUnavailable.

func TestPartyManagerReader_IsMemberTrueOn200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/teams/team-1/members/player-2" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reader := membership.NewPartyManagerReader(server.URL)
	member, err := reader.IsMember("team-1", "player-2")

	require.NoError(t, err)
	require.True(t, member, "expected membership confirmed on 200")
}

func TestPartyManagerReader_NotMemberOn404(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	reader := membership.NewPartyManagerReader(server.URL)
	member, err := reader.IsMember("team-1", "stranger")

	require.NoError(t, err)
	require.False(t, member, "expected non-member on 404")
}

func TestPartyManagerReader_UnavailableOn5xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	reader := membership.NewPartyManagerReader(server.URL)
	member, err := reader.IsMember("team-1", "player-2")

	require.ErrorIs(t, err, domain.ErrMembershipUnavailable)
	require.False(t, member, "membership must fail closed on 5xx")
}

func TestPartyManagerReader_UnavailableOnUnreachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // unreachable

	reader := membership.NewPartyManagerReader(server.URL)
	_, err := reader.IsMember("team-1", "player-2")

	require.ErrorIs(t, err, domain.ErrMembershipUnavailable)
}

func TestPartyManagerReader_UnavailableOnSlowResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reader := membership.NewPartyManagerReader(server.URL)
	_, err := reader.IsMember("team-1", "player-2")

	require.ErrorIs(t, err, domain.ErrMembershipUnavailable)
}

func TestPartyManagerReader_EscapesPathSegments(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	reader := membership.NewPartyManagerReader(server.URL)
	_, err := reader.IsMember("team/evil", "player?id=1")

	require.NoError(t, err)
	want := fmt.Sprintf("/teams/%s/members/%s", "team%2Fevil", "player%3Fid=1")
	require.Equal(t, want, gotPath)
}

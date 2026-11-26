package membership

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

const partyManagerTimeout = 2 * time.Second

// PartyManagerReader verifies team membership against the Party Manager HTTP
// API: GET {base}/teams/{team_id}/members/{player_id}. A 200 confirms
// membership, a 404 is a definitive negative, and every other outcome —
// including transport errors and timeouts — fails closed as
// domain.ErrMembershipUnavailable.
type PartyManagerReader struct {
	baseURL string
	client  *http.Client
}

// NewPartyManagerReader builds a reader against the given base URL with a
// fail-fast timeout so an unreachable source cannot hang the accept path.
func NewPartyManagerReader(baseURL string) *PartyManagerReader {
	return &PartyManagerReader{
		baseURL: baseURL,
		client:  &http.Client{Timeout: partyManagerTimeout},
	}
}

// IsMember reports whether playerID belongs to teamID.
func (r *PartyManagerReader) IsMember(teamID, playerID string) (bool, error) {
	endpoint := fmt.Sprintf("%s/teams/%s/members/%s",
		r.baseURL, url.PathEscape(teamID), url.PathEscape(playerID))
	resp, err := r.client.Get(endpoint)
	if err != nil {
		return false, fmt.Errorf("party manager membership check: %w", domain.ErrMembershipUnavailable)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("party manager status %d: %w", resp.StatusCode, domain.ErrMembershipUnavailable)
	}
}

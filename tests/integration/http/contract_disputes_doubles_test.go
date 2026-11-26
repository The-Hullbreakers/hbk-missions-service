package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// disputedContractFixture builds a started contract with terms so a member can
// raise a dispute.
func disputedContractFixture() *domain.MissionContract {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	terms, _ := domain.NewContractTerms([]domain.Objective{
		domain.NewObjectiveNamed("obj-a", "a", "", "BOUNTY", 10, 0, 0),
	})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	return contract
}

type disputeStore struct {
	contract *domain.MissionContract
}

func (s *disputeStore) Get(string) (*domain.MissionContract, error) { return s.contract, nil }
func (s *disputeStore) Save(c *domain.MissionContract) error {
	s.contract = c
	return nil
}

func (s *disputeStore) SaveDeparture(c *domain.MissionContract, _ string, _ domain.ContractState) error {
	return s.Save(c)
}

// memberSession authenticates as player-1 with no privileged roles.
type memberSession struct{}

func (memberSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "player-1", map[string]any{}, nil
}
func (memberSession) HasRole(map[string]any, string) bool { return false }

// moderatorSession authenticates as mod-1 with the moderator role.
type moderatorSession struct{}

func (moderatorSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "mod-1", map[string]any{"roles": []any{"moderator"}}, nil
}
func (moderatorSession) HasRole(claims map[string]any, role string) bool {
	roles, _ := claims["roles"].([]any)
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func postJSON(t *testing.T, server *httptest.Server, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, server.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

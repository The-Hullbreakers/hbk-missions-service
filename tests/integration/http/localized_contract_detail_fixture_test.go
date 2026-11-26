//go:build integration

package http_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// langMemberSession is a UserSession double admitting a plain roster member
// (no roles) so detail tests exercise the member projection path.
type langMemberSession struct{}

func (langMemberSession) GetUserIDFromHeader(string) (string, map[string]any, error) {
	return "member-1", map[string]any{"roles": []any{}}, nil
}

func (langMemberSession) HasRole(map[string]any, string) bool { return false }

// languageContractMemberServer boots the real composition server as member-1
// and seeds a bilingual contract and a legacy contract rostered to member-1
// plus one outsider contract that must stay forbidden.
func languageContractMemberServer(t *testing.T) (*httptest.Server, error) {
	return languageSessionServer(t, langMemberSession{}, func(client *missiondynamo.DynamoDBClient, table string) error {
		store := missiondynamo.NewContractStore(client, table)
		bilingual, err := bilingualContractFixture([]string{"member-1"})
		if err != nil {
			return err
		}
		if err := store.Save(bilingual); err != nil {
			return err
		}
		legacy, err := legacyContractFixture([]string{"member-1"})
		if err != nil {
			return err
		}
		if err := store.Save(legacy); err != nil {
			return err
		}
		outsider := domain.NewMissionContract("contract-foreign", legacy.Offer, []string{"player-9"})
		outsider.Terms = legacy.Terms
		return store.Save(outsider)
	})
}

// detailLangGet performs the authenticated contract-detail GET; errors
// surface to the test's Assert.
func detailLangGet(ts *httptest.Server, contractID, query, accept string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/missions/contracts/"+contractID+query, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer any")
	if accept != "" {
		request.Header.Set("Accept-Language", accept)
	}
	return ts.Client().Do(request)
}

// decodeDetail reads the response body into a generic detail map.
func decodeDetail(response *http.Response) (map[string]any, error) {
	var body map[string]any
	err := json.NewDecoder(response.Body).Decode(&body)
	return body, err
}

// detailTerm extracts the single frozen term objective from a detail body.
func detailTerm(body map[string]any) (map[string]any, error) {
	terms, ok := body["terms"].([]any)
	if !ok || len(terms) != 1 {
		return nil, fmt.Errorf("expected one contract term")
	}
	term, ok := terms[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected contract term object")
	}
	return term, nil
}

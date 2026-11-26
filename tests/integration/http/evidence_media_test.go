package http_test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

// evidenceBlobFake implements both evidence blob ports for boundary tests.
type evidenceBlobFake struct {
	exists  map[string]bool
	presign bool
}

func (f *evidenceBlobFake) PresignUpload(key, _ string, _ int64) (application.UploadGrant, error) {
	f.presign = true
	return application.UploadGrant{URL: "https://signed.example/upload?key=" + key, MediaURI: domain.MediaURI(key)}, nil
}

func (f *evidenceBlobFake) Exists(uri domain.MediaURI) (bool, error) {
	return f.exists[string(uri)], nil
}

func (f *evidenceBlobFake) PresignRead(uri domain.MediaURI) (application.ReadGrant, error) {
	return application.ReadGrant{URL: "https://signed.example/read?key=" + string(uri)}, nil
}

// roleSessionFake is a session double whose roles are configurable.
type roleSessionFake struct {
	playerID string
	roles    map[string]bool
}

func (s roleSessionFake) GetUserIDFromHeader(string) (string, map[string]any, error) {
	roles := make([]interface{}, 0, len(s.roles))
	for role := range s.roles {
		roles = append(roles, role)
	}
	return s.playerID, map[string]any{"roles": roles}, nil
}

func (s roleSessionFake) HasRole(_ map[string]any, role string) bool { return s.roles[role] }

func inProgressContract() *domain.MissionContract {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	objective := domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0)
	terms, _ := domain.NewContractTerms([]domain.Objective{objective})
	contract.Terms = terms
	member, _ := domain.NewCallerIdentity("player-1", nil, nil)
	_ = contract.Start(member)
	return contract
}

func newMediaTestServer(session roleSessionFake, contract *domain.MissionContract, blobs *evidenceBlobFake) *httptest.Server {
	upload := application.NewRequestEvidenceUploadHandler(fixedContractStore{contract: contract}, blobs, 50<<20)
	read := application.NewRequestEvidenceReadHandler(fixedContractStore{contract: contract}, blobs)
	auth := missionhttp.NewAuthMiddleware(session)

	mux := http.NewServeMux()
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/evidence/upload-url",
		auth.Authenticate(missionhttp.RequestEvidenceUploadHandler(upload)))
	mux.Handle("POST /api/v1/missions/contracts/{contract_id}/evidence/media-url",
		auth.Authenticate(missionhttp.RequestEvidenceReadHandler(read)))
	return httptest.NewServer(mux)
}

func postTo(t *testing.T, server *httptest.Server, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, server.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer any")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

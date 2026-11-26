package http_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestUploadURLEndpoint_GrantIssued(t *testing.T) {
	blobs := &evidenceBlobFake{}
	server := newMediaTestServer(roleSessionFake{playerID: "player-1"}, inProgressContract(), blobs)
	defer server.Close()

	resp := postTo(t, server,
		"/api/v1/missions/contracts/contract-1/evidence/upload-url",
		`{"content_type":"image/png","content_length":2048}`)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var body map[string]string
	err := json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	require.False(t, body["upload_url"] == "" || !strings.HasPrefix(body["media_uri"], "evidence/contract-1/"))
}

func TestUploadURLEndpoint_RejectsBadType(t *testing.T) {
	blobs := &evidenceBlobFake{}
	server := newMediaTestServer(roleSessionFake{playerID: "player-1"}, inProgressContract(), blobs)
	defer server.Close()

	resp := postTo(t, server,
		"/api/v1/missions/contracts/contract-1/evidence/upload-url",
		`{"content_type":"application/x-msdownload","content_length":2048}`)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.False(t, blobs.presign, "expected no presign call for disallowed type")
}

func TestUploadURLEndpoint_RejectsOutsider(t *testing.T) {
	blobs := &evidenceBlobFake{}
	server := newMediaTestServer(roleSessionFake{playerID: "outsider-9"}, inProgressContract(), blobs)
	defer server.Close()

	resp := postTo(t, server,
		"/api/v1/missions/contracts/contract-1/evidence/upload-url",
		`{"content_type":"image/png","content_length":2048}`)
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUploadURLEndpoint_RejectsNotInProgress(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	server := newMediaTestServer(roleSessionFake{playerID: "player-1"}, contract, &evidenceBlobFake{})
	defer server.Close()

	resp := postTo(t, server,
		"/api/v1/missions/contracts/contract-1/evidence/upload-url",
		`{"content_type":"image/png","content_length":2048}`)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Equal(t, "CONTRACT_NOT_IN_PROGRESS", body["code"])
}

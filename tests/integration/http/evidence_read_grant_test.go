package http_test

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestMediaURLEndpoint_ReviewerGetsReadGrant(t *testing.T) {
	blobs := &evidenceBlobFake{exists: map[string]bool{"evidence/contract-1/abc": true}}
	session := roleSessionFake{playerID: "ops-1", roles: map[string]bool{"moderator": true}}
	server := newMediaTestServer(session, inProgressContract(), blobs)
	defer server.Close()

	resp := postTo(t, server,
		"/api/v1/missions/contracts/contract-1/evidence/media-url",
		`{"media_uri":"evidence/contract-1/abc"}`)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&body)
	require.Contains(t, body["download_url"], "evidence/contract-1/abc")
}

func TestMediaURLEndpoint_MemberDenied(t *testing.T) {
	blobs := &evidenceBlobFake{exists: map[string]bool{"evidence/contract-1/abc": true}}
	server := newMediaTestServer(roleSessionFake{playerID: "player-1"}, inProgressContract(), blobs)
	defer server.Close()

	resp := postTo(t, server,
		"/api/v1/missions/contracts/contract-1/evidence/media-url",
		`{"media_uri":"evidence/contract-1/abc"}`)
	defer resp.Body.Close()

	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missionhttp "github.com/hbk/hbk-missions-service/internal/missions/interface/http"
)

type stagedContractWriter struct {
	saved *domain.MissionContract
}

func (w *stagedContractWriter) Save(c *domain.MissionContract) error {
	w.saved = c
	return nil
}

func newEvidenceMediaServer(blobs *evidenceBlobFake) (*httptest.Server, *stagedContractWriter) {
	contract := inProgressContract()
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	writer := &stagedContractWriter{}

	app := application.NewConfirmObjectiveHandlerWithEvidence(
		fixedContractStore{contract: contract},
		writer,
		fixedSeasonStore{season: season},
		blobs,
	)

	auth := missionhttp.NewAuthMiddleware(roleSessionFake{playerID: "player-1"})
	return httptest.NewServer(auth.Authenticate(missionhttp.ConfirmEvidenceHandler(app))), writer
}

func TestConfirmEvidenceEndpoint_AttachesMediaURI(t *testing.T) {
	blobs := &evidenceBlobFake{exists: map[string]bool{"evidence/contract-1/abc-123": true}}
	server, _ := newEvidenceMediaServer(blobs)
	defer server.Close()

	body := `{"objective_id":"objective-1","unit_id":"target-1","season_id":"season-1","media_uri":"evidence/contract-1/abc-123"}`
	resp := postEvidence(t, server, body)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var credit domain.MeritCredit
	err := json.NewDecoder(resp.Body).Decode(&credit)
	require.NoError(t, err)
	require.Equal(t, "evidence/contract-1/abc-123", string(credit.MediaURI))
}

func TestConfirmEvidenceEndpoint_RejectsPhantomMedia(t *testing.T) {
	blobs := &evidenceBlobFake{exists: map[string]bool{}}
	server, writer := newEvidenceMediaServer(blobs)
	defer server.Close()

	body := `{"objective_id":"objective-1","unit_id":"target-1","season_id":"season-1","media_uri":"evidence/contract-1/ghost"}`
	resp := postEvidence(t, server, body)
	defer resp.Body.Close()

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	var bodyJSON map[string]string
	err := json.NewDecoder(resp.Body).Decode(&bodyJSON)
	require.NoError(t, err)
	require.Equal(t, "EVIDENCE_MEDIA_INVALID", bodyJSON["code"])
	require.Nil(t, writer.saved)
}

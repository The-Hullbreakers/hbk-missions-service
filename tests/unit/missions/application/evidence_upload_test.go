package application_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// fakeEvidenceStore is a hand-written double for the evidence blob ports.
type fakeEvidenceStore struct {
	exists    map[string]bool
	existsErr error
	uploadErr error
	readErr   error
	uploadKey string
	uploadCT  string
	uploadMax int64
	readURI   domain.MediaURI
}

func (f *fakeEvidenceStore) PresignUpload(key, contentType string, maxBytes int64) (application.UploadGrant, error) {
	f.uploadKey, f.uploadCT, f.uploadMax = key, contentType, maxBytes
	return application.UploadGrant{URL: "https://signed.example/upload", MediaURI: domain.MediaURI(key)}, f.uploadErr
}

func (f *fakeEvidenceStore) Exists(uri domain.MediaURI) (bool, error) {
	if f.existsErr != nil {
		return false, f.existsErr
	}
	return f.exists[string(uri)], nil
}

// storedEvidence returns a fake evidence store where the given contract's
// object keys are already uploaded, plus the matching media URIs.
func storedEvidence(contractID string, objects ...string) (*fakeEvidenceStore, []string) {
	store := &fakeEvidenceStore{exists: map[string]bool{}}
	uris := make([]string, 0, len(objects))
	for _, object := range objects {
		uri := "evidence/" + contractID + "/" + object
		store.exists[uri] = true
		uris = append(uris, uri)
	}
	return store, uris
}

func (f *fakeEvidenceStore) PresignRead(uri domain.MediaURI) (application.ReadGrant, error) {
	f.readURI = uri
	return application.ReadGrant{URL: "https://signed.example/read"}, f.readErr
}

func TestRequestEvidenceUploadHandler_MemberReceivesGrant(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{}
	handler := application.NewRequestEvidenceUploadHandler(
		fixedContractStore{contract: contract}, store, 50<<20)

	grant, err := handler.Handle(testIdentity(t, "player-1", nil), "contract-1", "image/png", 1024)

	require.NoError(t, err)
	require.False(t, !strings.HasPrefix(string(grant.MediaURI), "evidence/contract-1/"))
	require.False(t, store.uploadCT != "image/png" || store.uploadMax != 50<<20)
}

func TestRequestEvidenceUploadHandler_RejectsNotInProgress(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	store := &fakeEvidenceStore{}
	handler := application.NewRequestEvidenceUploadHandler(
		fixedContractStore{contract: contract}, store, 50<<20)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "contract-1", "image/png", 1024)

	require.ErrorIs(t, err, domain.ErrContractNotInProgress)
	require.Empty(t, store.uploadKey)
}

func TestRequestEvidenceUploadHandler_RejectsOutsider(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{}
	handler := application.NewRequestEvidenceUploadHandler(
		fixedContractStore{contract: contract}, store, 50<<20)

	_, err := handler.Handle(testIdentity(t, "player-2", nil), "contract-1", "image/png", 1024)

	require.ErrorIs(t, err, domain.ErrCallerNotAuthorized)
	require.Empty(t, store.uploadKey)
}

func TestRequestEvidenceUploadHandler_RejectsDeclaredOversize(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{}
	handler := application.NewRequestEvidenceUploadHandler(
		fixedContractStore{contract: contract}, store, 50<<20)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "contract-1", "image/png", 60<<20)

	require.ErrorIs(t, err, domain.ErrEvidenceUploadTooLarge)
	require.Empty(t, store.uploadKey)
}

func TestRequestEvidenceUploadHandler_PropagatesStoreError(t *testing.T) {
	handler := application.NewRequestEvidenceUploadHandler(
		failingContractStore{}, &fakeEvidenceStore{}, 50<<20)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "contract-1", "image/png", 1024)

	require.Error(t, err)
}

func TestRequestEvidenceUploadHandler_PropagatesPresignError(t *testing.T) {
	contract := newInProgressContract()
	broken := errors.New("presign failed")
	handler := application.NewRequestEvidenceUploadHandler(
		fixedContractStore{contract: contract}, &fakeEvidenceStore{uploadErr: broken}, 50<<20)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "contract-1", "image/png", 1024)

	require.ErrorIs(t, err, broken)
}

func TestRequestEvidenceUploadHandler_RejectsMediaType(t *testing.T) {
	contract := newInProgressContract()
	store := &fakeEvidenceStore{}
	handler := application.NewRequestEvidenceUploadHandler(
		fixedContractStore{contract: contract}, store, 50<<20)

	_, err := handler.Handle(testIdentity(t, "player-1", nil), "contract-1", "application/x-msdownload", 1024)

	require.ErrorIs(t, err, domain.ErrMediaTypeNotAllowed)
	require.Empty(t, store.uploadKey)
}

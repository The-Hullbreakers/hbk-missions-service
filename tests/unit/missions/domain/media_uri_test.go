package domain_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestMediaURI_AcceptsContractScopedReference(t *testing.T) {
	uri, err := domain.NewMediaURI("evidence/contract-1/abc-123", "contract-1")
	require.NoError(t, err)
	require.Equal(t, "evidence/contract-1/abc-123", string(uri))
}

func TestMediaURI_RejectsForeignContract(t *testing.T) {
	_, err := domain.NewMediaURI("evidence/contract-2/abc-123", "contract-1")
	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
}

func TestMediaURI_RejectsEmpty(t *testing.T) {
	_, err := domain.NewMediaURI("", "contract-1")
	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
}

func TestMediaURI_RejectsTraversal(t *testing.T) {
	_, err := domain.NewMediaURI("evidence/contract-1/../contract-2/x", "contract-1")
	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
}

func TestMediaURI_RejectsAbsoluteURL(t *testing.T) {
	_, err := domain.NewMediaURI("https://bucket.s3.amazonaws.com/evidence/contract-1/x", "contract-1")
	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
}

func TestMediaURI_RejectsMissingObjectID(t *testing.T) {
	_, err := domain.NewMediaURI("evidence/contract-1/", "contract-1")
	require.ErrorIs(t, err, domain.ErrEvidenceMediaInvalid)
}

func TestMissionContract_MediaPrefix(t *testing.T) {
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})

	require.Equal(t, "evidence/contract-1/", string(contract.MediaPrefix()))
}

func TestAllowedMediaType_AcceptsImageAndVideo(t *testing.T) {
	for _, contentType := range []string{"image/png", "image/jpeg", "video/mp4", "video/webm"} {
		require.True(t, domain.AllowedMediaType(contentType), "expected %s to be allowed", contentType)
	}
}

func TestAllowedMediaType_RejectsOtherTypes(t *testing.T) {
	for _, contentType := range []string{"", "application/x-msdownload", "text/html", "image", "video"} {
		require.False(t, domain.AllowedMediaType(contentType), "expected %q to be rejected", contentType)
	}
}

func TestMissionContract_ConfirmWithMedia_StoresReference(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	freezeTerms(t, contract, domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0))
	startContract(t, contract)

	media, err := domain.NewMediaURI("evidence/contract-1/abc-123", "contract-1")
	require.NoError(t, err)
	credit, err := contract.ConfirmWithMedia("objective-1", "target-1", media, season)
	require.NoError(t, err)
	require.Equal(t, "evidence/contract-1/abc-123", string(credit.MediaURI))
}

func TestMissionContract_Confirm_LeavesMediaEmpty(t *testing.T) {
	season := domain.NewSeason("season-1", domain.SeasonStateActive)
	offer := domain.NewMissionOffer("offer-1", "BOUNTY_HUNTING", "season-1")
	contract := domain.NewMissionContract("contract-1", offer, []string{"player-1"})
	freezeTerms(t, contract, domain.NewObjectiveNamed("objective-1", "objective-1", "", "BOUNTY_HUNTING", 100, 0, 0))
	startContract(t, contract)

	credit, err := contract.Confirm("objective-1", "target-1", season)
	require.NoError(t, err)
	require.Empty(t, credit.MediaURI)
}

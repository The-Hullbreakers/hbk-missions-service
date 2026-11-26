package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// TEST-DOM-ADJ-003 — Digest is the canonical per-fact content hash the
// staged set aggregates over; semantics asserted, opaque bytes never.
func TestPlacementFact_DigestChangesWithContent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		mutate func(*domain.PlacementFactParams)
	}{
		{"final merits", func(p *domain.PlacementFactParams) { p.FinalMerits = 301 }},
		{"member identity", func(p *domain.PlacementFactParams) { p.MemberIDs = []string{"m-1", "m-9"} }},
		{"member added", func(p *domain.PlacementFactParams) { p.MemberIDs = []string{"m-1", "m-2", "m-3"} }},
		{"season", func(p *domain.PlacementFactParams) { p.SeasonID = "season-2" }},
		{"category", func(p *domain.PlacementFactParams) { p.Category = "EXPLORATION" }},
		{"track", func(p *domain.PlacementFactParams) { p.Track = domain.TrackIndividual }},
		{"recipient", func(p *domain.PlacementFactParams) { p.RecipientID = "team-b" }},
		{"rank", func(p *domain.PlacementFactParams) { p.Rank = 2 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			base := mustAdjFact(t, adjFactParams())
			want := base.Digest()
			variantParams := adjFactParams()
			tc.mutate(&variantParams)
			variant := mustAdjFact(t, variantParams)

			// Act
			got := variant.Digest()

			// Assert
			require.NotEmpty(t, got)
			assert.NotEqual(t, want, got, "%s must feed the canonical digest", tc.name)
		})
	}
}

func TestPlacementFact_DigestMemberOrderInsensitive(t *testing.T) {
	t.Parallel()

	// Arrange — roster ordering is presentation, not fact content.
	want := mustAdjFact(t, adjFactParams()).Digest()
	variantParams := adjFactParams()
	variantParams.MemberIDs = []string{"m-2", "m-1"}
	variant := mustAdjFact(t, variantParams)

	// Act
	got := variant.Digest()

	// Assert
	assert.Equal(t, want, got, "member ordering alone must not change the digest")
}

func TestPlacementFact_DigestStableAcrossReads(t *testing.T) {
	t.Parallel()

	// Arrange
	fact := mustAdjFact(t, adjFactParams())
	want := fact.Digest()

	// Act
	got := fact.Digest()

	// Assert
	require.NotEmpty(t, want)
	assert.Equal(t, want, got)
}

func TestPlacementFact_DigestEventIDParticipates(t *testing.T) {
	t.Parallel()

	// Arrange — persisted event identity is part of the staged canonical
	// content; a replay carrying foreign identity must not pass as equal.
	want := mustAdjFact(t, adjFactParams()).Digest()
	variant := mustAdjFact(t, adjFactParams())
	variant.EventID = "evt-foreign-replay"

	// Act
	got := variant.Digest()

	// Assert
	assert.NotEqual(t, want, got, "EventID must feed the canonical digest")
}

func TestPlacementFact_DigestDistinctFactsDistinctDigests(t *testing.T) {
	t.Parallel()

	// Arrange — same bucket, different rank: two separate official entries.
	p := adjFactParams()
	first := mustAdjFact(t, p)
	p.Rank, p.RecipientID = 2, "team-b"
	second := mustAdjFact(t, p)
	want := first.Digest()

	// Act
	got := second.Digest()

	// Assert
	assert.NotEqual(t, want, got)
}

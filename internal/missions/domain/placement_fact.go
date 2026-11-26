package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"strconv"
)

// PlacementFactParams are the inputs required to build a placement fact.
type PlacementFactParams struct {
	SeasonID    string
	Category    string
	Track       Track
	RecipientID string
	Rank        int
	FinalMerits int
	MemberIDs   []string
}

// PlacementFact is the immutable business fact that one ranked entry was
// awarded a final position in a finalized official season.
type PlacementFact struct {
	EventID     string
	SeasonID    string
	Category    string
	Track       Track
	RecipientID string
	Rank        int
	FinalMerits int
	MemberIDs   []string
}

// NewPlacementFact validates the inputs and derives a deterministic event ID
// so that re-emitting the same placement produces the same fact identity.
func NewPlacementFact(p PlacementFactParams) (PlacementFact, error) {
	if p.SeasonID == "" || p.Category == "" || p.RecipientID == "" || p.Rank < 1 || p.FinalMerits < 0 {
		return PlacementFact{}, ErrPlacementFactInvalid
	}
	if p.Track == TrackTeam && len(p.MemberIDs) == 0 {
		return PlacementFact{}, ErrTeamFactRequiresMembers
	}
	fact := PlacementFact{
		SeasonID:    p.SeasonID,
		Category:    p.Category,
		Track:       p.Track,
		RecipientID: p.RecipientID,
		Rank:        p.Rank,
		FinalMerits: p.FinalMerits,
		MemberIDs:   slices.Clone(p.MemberIDs),
	}
	fact.EventID = placementEventID(fact)
	return fact, nil
}

// Digest is the canonical per-fact content hash the staged fact set
// aggregates over. Every placement-relevant field participates — including
// the persisted EventID so a replay carrying foreign identity can never
// pass as equal — and member ordering alone cannot change the digest
// while member multiplicity is preserved.
func (f PlacementFact) Digest() string {
	members := slices.Clone(f.MemberIDs)
	sort.Strings(members)
	fields := []string{
		f.EventID, f.SeasonID, f.Category, string(f.Track),
		f.RecipientID, strconv.Itoa(f.Rank), strconv.Itoa(f.FinalMerits),
	}
	return hashFields(append(fields, members...))
}

func placementEventID(f PlacementFact) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%s|%d", f.SeasonID, f.Category, f.Track, f.RecipientID, f.Rank)))
	return hex.EncodeToString(sum[:16])
}

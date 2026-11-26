package domain

import "sort"

// StagedFactSet is the immutable accounting of the placement facts staged
// for one adjudication attempt. It keeps only the declared count and the
// aggregate hash captured at construction, so later mutations of the source
// facts or their member slices can never rewrite what was staged.
type StagedFactSet struct {
	count int
	hash  string
}

// NewStagedFactSet captures count and aggregate hash over the canonical
// per-fact digests. Aggregation is order-independent: digests are sorted
// before hashing. Every fact must carry its persisted event identity, and
// duplicate identities are rejected — replay can never silently restage a
// different fact under an existing EventID.
func NewStagedFactSet(facts []PlacementFact) (StagedFactSet, error) {
	seen := make(map[string]struct{}, len(facts))
	digests := make([]string, 0, len(facts))
	for _, f := range facts {
		if f.EventID == "" {
			return StagedFactSet{}, ErrStagedFactSetInvalid
		}
		if _, dup := seen[f.EventID]; dup {
			return StagedFactSet{}, ErrStagedFactSetInvalid
		}
		seen[f.EventID] = struct{}{}
		digests = append(digests, f.Digest())
	}
	sort.Strings(digests)
	return StagedFactSet{count: len(facts), hash: hashFields(digests)}, nil
}

func (s StagedFactSet) Count() int { return s.count }

func (s StagedFactSet) Hash() string { return s.hash }

// Verify is the commit gatekeeper: the manifest may commit only when the
// declared count and hash match exactly what was staged.
func (s StagedFactSet) Verify(count int, hash string) error {
	if count != s.count || hash != s.hash {
		return ErrStagedFactSetMismatch
	}
	return nil
}

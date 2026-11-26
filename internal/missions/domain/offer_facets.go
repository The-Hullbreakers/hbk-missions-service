package domain

import "sort"

// OfferFacets lists the distinct category and mode values present.
type OfferFacets struct {
	Categories []string      `json:"categories"`
	Modes      []MissionMode `json:"modes"`
}

// CollectFacets returns the distinct, sorted categories and modes in offers.
func CollectFacets(offers []MissionOffer) OfferFacets {
	categories := map[string]bool{}
	modes := map[MissionMode]bool{}
	for _, o := range offers {
		if o.Category != "" {
			categories[o.Category] = true
		}
		if o.Mode != "" {
			modes[o.Mode] = true
		}
	}
	facets := OfferFacets{Categories: []string{}, Modes: []MissionMode{}}
	for c := range categories {
		facets.Categories = append(facets.Categories, c)
	}
	for m := range modes {
		facets.Modes = append(facets.Modes, m)
	}
	sort.Strings(facets.Categories)
	sort.Slice(facets.Modes, func(i, j int) bool { return facets.Modes[i] < facets.Modes[j] })
	return facets
}

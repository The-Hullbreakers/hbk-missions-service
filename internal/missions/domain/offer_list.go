package domain

import "strings"

// OfferStatusFilter is a business bucket for listing offers.
type OfferStatusFilter string

const (
	// OfferStatusAvailable lists offers open for acceptance (the default).
	OfferStatusAvailable OfferStatusFilter = "AVAILABLE"
	// OfferStatusAll lists every published offer regardless of availability.
	OfferStatusAll OfferStatusFilter = "ALL"
)

// OfferListQuery carries the raw HTTP query parameters for listing offers.
type OfferListQuery struct {
	SeasonID string
	Status   string
	Category string
	Mode     string
	Q        string
	Limit    string
	Offset   string
	Sort     string
}

// OfferListFilter is the parsed, typed query for the mission board.
// Status is resolved by the reader (availability depends on the offer window);
// Match applies the structural fields only.
type OfferListFilter struct {
	SeasonID   string
	Status     OfferStatusFilter
	Categories []string
	Modes      []MissionMode
	Q          string
	Page       OfferPage
	Sort       OfferSort
}

// ParseOfferListFilter validates the raw query into a typed filter.
// Empty status defaults to AVAILABLE; category/mode accept comma-separated
// lists with OR semantics. Unknown values return an error.
func ParseOfferListFilter(raw OfferListQuery) (OfferListFilter, error) {
	filter := OfferListFilter{
		SeasonID:   raw.SeasonID,
		Status:     OfferStatusAvailable,
		Categories: parseCSV(raw.Category),
		Q:          raw.Q,
	}
	modes, err := parseModes(raw.Mode)
	if err != nil {
		return OfferListFilter{}, err
	}
	filter.Modes = modes
	switch strings.ToUpper(strings.TrimSpace(raw.Status)) {
	case "", string(OfferStatusAvailable):
	case string(OfferStatusAll):
		filter.Status = OfferStatusAll
	default:
		return OfferListFilter{}, ErrInvalidOfferStatusFilter
	}
	page, err := ParseOfferPage(raw.Limit, raw.Offset)
	if err != nil {
		return OfferListFilter{}, err
	}
	filter.Page = page
	sort, err := ParseOfferSort(raw.Sort)
	if err != nil {
		return OfferListFilter{}, err
	}
	filter.Sort = sort
	return filter, nil
}

// Match reports whether the offer satisfies the structural fields.
func (f OfferListFilter) Match(offer MissionOffer) bool {
	if f.SeasonID != "" && offer.SeasonID != f.SeasonID {
		return false
	}
	if len(f.Categories) > 0 && !containsString(f.Categories, offer.Category) {
		return false
	}
	if len(f.Modes) > 0 && !containsMode(f.Modes, offer.Mode) {
		return false
	}
	if f.Q != "" && !matchesQuery(offer, f.Q) {
		return false
	}
	return true
}

// matchesQuery reports a case-insensitive substring match over title and
// description.
func matchesQuery(offer MissionOffer, q string) bool {
	text := strings.ToLower(offer.Title + " " + offer.Description)
	return strings.Contains(text, strings.ToLower(q))
}

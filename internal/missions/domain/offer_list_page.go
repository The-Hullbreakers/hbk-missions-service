package domain

import (
	"sort"
	"strings"
)

const (
	// OfferListDefaultLimit bounds pages when the caller omits limit.
	OfferListDefaultLimit = 50
	// OfferListMaxLimit caps any requested page size.
	OfferListMaxLimit = 200
)

// OfferPage carries validated pagination bounds.
type OfferPage struct {
	Limit  int
	Offset int
}

// ParseOfferPage validates raw limit/offset parameters.
func ParseOfferPage(rawLimit, rawOffset string) (OfferPage, error) {
	limit, err := parseBoundedInt(rawLimit, OfferListDefaultLimit, 1, OfferListMaxLimit)
	if err != nil {
		return OfferPage{}, err
	}
	offset, err := parseBoundedInt(rawOffset, 0, 0, 0)
	if err != nil {
		return OfferPage{}, err
	}
	return OfferPage{Limit: limit, Offset: offset}, nil
}

// OfferSortField is a whitelisted sortable offer attribute.
type OfferSortField string

const (
	// OfferSortCompletionReward orders by the frozen reward amount.
	OfferSortCompletionReward OfferSortField = "completion_reward"
	// OfferSortCategory orders alphabetically by category.
	OfferSortCategory OfferSortField = "category"
)

// OfferSort is the parsed sort directive; a leading "-" means descending.
type OfferSort struct {
	Field      OfferSortField
	Descending bool
}

// ParseOfferSort validates the raw sort parameter.
func ParseOfferSort(raw string) (OfferSort, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return OfferSort{}, nil
	}
	desc := strings.HasPrefix(raw, "-")
	field := OfferSortField(strings.TrimPrefix(raw, "-"))
	switch field {
	case OfferSortCompletionReward, OfferSortCategory:
		return OfferSort{Field: field, Descending: desc}, nil
	default:
		return OfferSort{}, ErrInvalidOfferListFilter
	}
}

// OfferListPage is the paginated list response envelope.
type OfferListPage struct {
	Offers []MissionOffer `json:"offers"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// SortOffers orders offers by the whitelisted field; ties and the default
// keep a stable OfferID ascending order.
func SortOffers(offers []MissionOffer, s OfferSort) []MissionOffer {
	out := make([]MissionOffer, len(offers))
	copy(out, offers)
	sort.SliceStable(out, func(i, j int) bool {
		if cmp := compareOffers(out[i], out[j], s.Field); cmp != 0 {
			if s.Descending {
				return cmp > 0
			}
			return cmp < 0
		}
		return out[i].OfferID < out[j].OfferID
	})
	return out
}

func compareOffers(a, b MissionOffer, field OfferSortField) int {
	switch field {
	case OfferSortCompletionReward:
		return a.CompletionReward - b.CompletionReward
	case OfferSortCategory:
		return strings.Compare(a.Category, b.Category)
	default:
		return strings.Compare(a.OfferID, b.OfferID)
	}
}

// PaginateOffers slices the sorted, visible offers into a page.
func PaginateOffers(offers []MissionOffer, page OfferPage) OfferListPage {
	total := len(offers)
	result := OfferListPage{Offers: []MissionOffer{}, Total: total, Limit: page.Limit, Offset: page.Offset}
	if page.Offset >= total {
		return result
	}
	end := page.Offset + page.Limit
	if page.Limit <= 0 || end > total {
		end = total
	}
	result.Offers = offers[page.Offset:end]
	return result
}

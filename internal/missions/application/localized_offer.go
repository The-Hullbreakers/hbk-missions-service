package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// LocalizedOffer is the offer read projection in the language the caller
// selected: the offer fields keep their shape while ContentLanguage
// honestly reports the locale actually served.
type LocalizedOffer struct {
	domain.MissionOffer
	ContentLanguage domain.Language
}

// LocalizedOfferListPage is the paginated mission-board envelope in the
// caller's language: per-item ContentLanguage is authoritative for mixed
// pages. Envelope keys mirror the existing offer list response.
type LocalizedOfferListPage struct {
	Offers []LocalizedOffer `json:"offers"`
	Total  int              `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

// LocalizeOffer projects an offer into the requested language via the
// shared atomic text resolution; the stored offer is never mutated.
func LocalizeOffer(offer domain.MissionOffer, language domain.Language) LocalizedOffer {
	text := localizeMissionText(offer.Title, offer.Description, offer.Rules, offer.Objectives, offer.ContentEn, language)
	served := LocalizedOffer{MissionOffer: offer, ContentLanguage: text.Language}
	served.Title = text.Title
	served.Description = text.Description
	served.Rules = text.Rules
	served.Objectives = text.Objectives
	return served
}

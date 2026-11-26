package application

import "github.com/hbk/hbk-missions-service/internal/missions/domain"

// LocalizedContract is the contract read projection in the language the
// caller selected: one ContentLanguage honestly reports the locale served
// for the embedded offer and frozen terms together.
type LocalizedContract struct {
	domain.MissionContract
	ContentLanguage domain.Language
}

// LocalizeContract projects a contract into the requested language via the
// shared atomic text resolution. The effective objective set is the frozen
// Terms snapshot when present — legacy persisted offers may carry no
// objectives — otherwise the offer's own list. English is served only when
// the snapshot covers that set; nonzero terms then get their own
// translated snapshot. Anything unexpected falls the whole contract back
// to Spanish; the stored contract is never mutated.
func LocalizeContract(contract domain.MissionContract, language domain.Language) LocalizedContract {
	effective := contract.Offer.Objectives
	if !contract.Terms.IsZero() {
		effective = contract.Terms.Objectives()
	}
	text := localizeMissionText(contract.Offer.Title, contract.Offer.Description,
		contract.Offer.Rules, effective, contract.Offer.ContentEn, language)
	served := LocalizedContract{MissionContract: contract, ContentLanguage: text.Language}
	served.Offer.Title = text.Title
	served.Offer.Description = text.Description
	served.Offer.Rules = text.Rules
	served.Offer.Objectives = text.Objectives
	if text.Language != domain.LanguageEnglish || contract.Terms.IsZero() {
		return served
	}
	translated, err := contract.Terms.WithObjectiveText(contract.Offer.ContentEn.Objectives)
	if err != nil {
		return LocalizedContract{MissionContract: contract, ContentLanguage: domain.LanguageSpanish}
	}
	served.Terms = translated
	return served
}

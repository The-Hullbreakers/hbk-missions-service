package http

import (
	nethttp "net/http"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// requestedMissionLanguage validates one mission-content language query.
// Omitted lang defaults to Spanish; repeated values must agree.
func requestedMissionLanguage(r *nethttp.Request) (domain.Language, error) {
	values, present := r.URL.Query()["lang"]
	if !present {
		return domain.LanguageSpanish, nil
	}
	language, err := domain.ParseLanguage(values[0])
	if err != nil {
		return "", err
	}
	for _, value := range values[1:] {
		if value != string(language) {
			return "", domain.ErrLanguageInvalid
		}
	}
	return language, nil
}

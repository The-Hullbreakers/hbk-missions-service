package domain

import "errors"

// Language identifies a supported mission-content locale. Spanish is the
// default and the historical baseline; English is the first translation.
type Language string

const (
	// LanguageSpanish is the default served language.
	LanguageSpanish Language = "es"
	// LanguageEnglish is the English translation locale.
	LanguageEnglish Language = "en"
)

// ErrLanguageInvalid reports an empty or unsupported language code.
var ErrLanguageInvalid = errors.New("language is invalid")

// ParseLanguage validates a supported language code. Transport repetition
// syntax never reaches this parser.
func ParseLanguage(raw string) (Language, error) {
	switch Language(raw) {
	case LanguageSpanish:
		return LanguageSpanish, nil
	case LanguageEnglish:
		return LanguageEnglish, nil
	default:
		return "", ErrLanguageInvalid
	}
}

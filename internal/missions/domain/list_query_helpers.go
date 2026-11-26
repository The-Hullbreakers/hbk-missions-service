package domain

import (
	"strconv"
	"strings"
)

// parseCSV splits a comma-separated list, upper-casing and trimming entries
// and dropping empty segments.
func parseCSV(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.ToUpper(strings.TrimSpace(p)); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// parseModes validates a CSV of mission modes.
func parseModes(raw string) ([]MissionMode, error) {
	values := parseCSV(raw)
	modes := make([]MissionMode, 0, len(values))
	for _, v := range values {
		mode := MissionMode(v)
		if !mode.IsValid() {
			return nil, ErrInvalidOfferModeFilter
		}
		modes = append(modes, mode)
	}
	return modes, nil
}

func containsString(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func containsMode(list []MissionMode, value MissionMode) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

// parseBoundedInt validates a numeric query parameter: empty returns def,
// values below min or non-numeric fail, and max > 0 clamps the upper bound.
func parseBoundedInt(raw string, def, min, max int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < min {
		return 0, ErrInvalidOfferListFilter
	}
	if max > 0 && n > max {
		return max, nil
	}
	return n, nil
}

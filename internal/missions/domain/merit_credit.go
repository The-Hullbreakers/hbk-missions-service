package domain

import (
	"errors"
	"strings"
)

// Track identifies the competitive track of a credit.
type Track string

const (
	TrackIndividual Track = "INDIVIDUAL"
	TrackTeam       Track = "TEAM"
)

// ErrInvalidTrack is returned when a track value is not recognized.
var ErrInvalidTrack = errors.New("invalid track")

// ParseTrack converts a raw string to a Track. Unknown values return ErrInvalidTrack.
func ParseTrack(raw string) (Track, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(TrackIndividual):
		return TrackIndividual, nil
	case string(TrackTeam):
		return TrackTeam, nil
	default:
		return "", ErrInvalidTrack
	}
}

// MeritCredit is an immutable record of merits granted for a confirmed objective.
type MeritCredit struct {
	SeasonID       string
	Category       string
	Track          Track
	RecipientID    string
	Amount         int
	ContractID     string
	ObjectiveID    string
	UnitID         string
	ConfirmationID string
	// MediaURI references the stored evidence object, when one was attached.
	MediaURI MediaURI
}

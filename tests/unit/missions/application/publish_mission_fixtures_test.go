package application_test

import (
	"errors"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type recordingOfferWriter struct {
	saved []domain.MissionOffer
}

func (w *recordingOfferWriter) Save(offer domain.MissionOffer) error {
	w.saved = append(w.saved, offer)
	return nil
}

type failingOfferWriter struct{}

func (failingOfferWriter) Save(domain.MissionOffer) error {
	return errors.New("offer writer unavailable")
}

type fixedActiveSeasonsStore struct {
	seasons []domain.Season
	err     error
}

func (s fixedActiveSeasonsStore) ListActive() ([]domain.Season, error) {
	return s.seasons, s.err
}

func anActiveSeason() fixedActiveSeasonsStore {
	return fixedActiveSeasonsStore{seasons: []domain.Season{{SeasonID: "season-2026-q4", State: domain.SeasonStateActive, Official: true}}}
}

type fixedDefinitionReader struct {
	definition domain.MissionDefinition
	err        error
}

func (r fixedDefinitionReader) Get(_ string) (domain.MissionDefinition, error) {
	return r.definition, r.err
}

func (r fixedDefinitionReader) GetVersion(_, _ string) (domain.MissionDefinition, error) {
	return r.definition, r.err
}

func aPublishedDefinition() domain.MissionDefinition {
	def, _ := domain.NewMissionDefinition("bounty-template", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	def.ContentEn = domain.LocalizedContent{Title: "Bounty", Description: "Desc"}
	_ = def.Publish("v1")
	return def
}

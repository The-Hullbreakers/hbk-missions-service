package application_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestPublishMissionHandler_PublishesFromDefinition(t *testing.T) {
	writer := &recordingOfferWriter{}
	handler := application.NewPublishMissionHandler(writer, fixedDefinitionReader{definition: aPublishedDefinition()}, anActiveSeason())

	offer, err := handler.Handle("offer-1", "bounty-template", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if offer.OfferID != "offer-1" || offer.Mode != domain.MissionModeSolo {
		t.Fatalf("expected solo offer-1, got %+v", offer)
	}
	if offer.SeasonID != "season-2026-q4" {
		t.Fatalf("expected season season-2026-q4, got %s", offer.SeasonID)
	}
	if offer.DefinitionCode != "bounty-template" {
		t.Fatalf("expected definition bounty-template, got %s", offer.DefinitionCode)
	}
	if len(writer.saved) != 1 {
		t.Fatalf("expected offer to be saved")
	}
}

func TestPublishMissionHandler_RejectsNoActiveSeason(t *testing.T) {
	writer := &recordingOfferWriter{}
	handler := application.NewPublishMissionHandler(writer, fixedDefinitionReader{definition: aPublishedDefinition()}, fixedActiveSeasonsStore{seasons: []domain.Season{}})

	_, err := handler.Handle("offer-1", "bounty-template", "v1")
	if !errors.Is(err, domain.ErrNoActiveSeason) {
		t.Fatalf("expected ErrNoActiveSeason, got %v", err)
	}
	if len(writer.saved) != 0 {
		t.Fatalf("expected no offer to be saved")
	}
}

func TestPublishMissionHandler_RejectsUnpublishedDefinition(t *testing.T) {
	writer := &recordingOfferWriter{}
	draft, _ := domain.NewMissionDefinition("bounty-template", "T", "D", "BOUNTY_HUNTING", domain.MissionModeSolo, 0, nil, nil)
	handler := application.NewPublishMissionHandler(writer, fixedDefinitionReader{definition: draft}, anActiveSeason())

	_, err := handler.Handle("offer-1", "bounty-template", "")
	if !errors.Is(err, domain.ErrDefinitionNotPublished) {
		t.Fatalf("expected ErrDefinitionNotPublished, got %v", err)
	}
}

func TestPublishMissionHandler_ReturnsWriterError(t *testing.T) {
	handler := application.NewPublishMissionHandler(failingOfferWriter{}, fixedDefinitionReader{definition: aPublishedDefinition()}, anActiveSeason())
	_, err := handler.Handle("offer-1", "bounty-template", "v1")
	if err == nil {
		t.Fatalf("expected writer error")
	}
}

func TestPublishMissionHandler_ReturnsSeasonsError(t *testing.T) {
	writer := &recordingOfferWriter{}
	handler := application.NewPublishMissionHandler(writer, fixedDefinitionReader{definition: aPublishedDefinition()}, fixedActiveSeasonsStore{err: errors.New("seasons unavailable")})

	_, err := handler.Handle("offer-1", "bounty-template", "v1")
	if err == nil {
		t.Fatalf("expected seasons error")
	}
}

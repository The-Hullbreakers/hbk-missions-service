package domain_test

import (
	"errors"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func TestBuildLeaderboard_SortsByMeritsDescending(t *testing.T) {
	// Arrange
	standings := []domain.Standing{
		{RecipientID: "player-2", Total: 100},
		{RecipientID: "player-1", Total: 150},
	}

	// Act
	board, err := domain.BuildLeaderboard(standings, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert
	if len(board.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(board.Entries))
	}
	if board.Entries[0].RecipientID != "player-1" || board.Entries[0].Rank != 1 {
		t.Fatalf("expected player-1 at rank 1, got %s at rank %d", board.Entries[0].RecipientID, board.Entries[0].Rank)
	}
	if board.Entries[1].RecipientID != "player-2" || board.Entries[1].Rank != 2 {
		t.Fatalf("expected player-2 at rank 2, got %s at rank %d", board.Entries[1].RecipientID, board.Entries[1].Rank)
	}
}

func TestBuildLeaderboard_TieBreakIsDeterministic(t *testing.T) {
	// Arrange
	standings := []domain.Standing{
		{RecipientID: "player-b", Total: 150, CompletedMissions: 2, UniqueObjectives: 3},
		{RecipientID: "player-a", Total: 150, CompletedMissions: 2, UniqueObjectives: 3},
	}

	// Act
	board, err := domain.BuildLeaderboard(standings, 10, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert
	if len(board.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(board.Entries))
	}
	if board.Entries[0].Rank != 1 || board.Entries[1].Rank != 2 {
		t.Fatalf("expected ranks 1 and 2, got %d and %d", board.Entries[0].Rank, board.Entries[1].Rank)
	}
}

func TestBuildLeaderboard_PaginatesWithCursor(t *testing.T) {
	// Arrange
	standings := []domain.Standing{
		{RecipientID: "player-1", Total: 500},
		{RecipientID: "player-2", Total: 400},
		{RecipientID: "player-3", Total: 300},
		{RecipientID: "player-4", Total: 200},
		{RecipientID: "player-5", Total: 100},
	}

	// Act first page
	first, err := domain.BuildLeaderboard(standings, 2, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert first page
	if len(first.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(first.Entries))
	}
	if first.Entries[0].RecipientID != "player-1" {
		t.Fatalf("expected player-1 first, got %s", first.Entries[0].RecipientID)
	}
	if first.NextCursor == "" {
		t.Fatalf("expected non-empty next cursor")
	}

	// Act second page
	second, err := domain.BuildLeaderboard(standings, 2, first.NextCursor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert second page
	if len(second.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(second.Entries))
	}
	if second.Entries[0].RecipientID != "player-3" {
		t.Fatalf("expected player-3 first on second page, got %s", second.Entries[0].RecipientID)
	}
}

func TestBuildLeaderboard_RejectsInvalidCursor(t *testing.T) {
	_, err := domain.BuildLeaderboard(nil, 10, "not-a-valid-cursor")
	if !errors.Is(err, domain.ErrInvalidCursor) {
		t.Fatalf("expected ErrInvalidCursor, got %v", err)
	}
}

func TestCompareStanding_CoversBranches(t *testing.T) {
	if domain.CompareStanding(
		domain.Standing{Total: 200},
		domain.Standing{Total: 100},
	) >= 0 {
		t.Fatalf("expected higher total to come first")
	}
	if domain.CompareStanding(
		domain.Standing{Total: 100, CompletedMissions: 5},
		domain.Standing{Total: 100, CompletedMissions: 3},
	) >= 0 {
		t.Fatalf("expected more completed missions to come first")
	}
	if domain.CompareStanding(
		domain.Standing{Total: 100, CompletedMissions: 3, UniqueObjectives: 5},
		domain.Standing{Total: 100, CompletedMissions: 3, UniqueObjectives: 3},
	) >= 0 {
		t.Fatalf("expected more unique objectives to come first")
	}
	if domain.CompareStanding(
		domain.Standing{RecipientID: "a"},
		domain.Standing{RecipientID: "b"},
	) >= 0 {
		t.Fatalf("expected a before b")
	}
	if domain.CompareStanding(
		domain.Standing{RecipientID: "b"},
		domain.Standing{RecipientID: "a"},
	) <= 0 {
		t.Fatalf("expected b after a")
	}
	if domain.CompareStanding(
		domain.Standing{RecipientID: "a"},
		domain.Standing{RecipientID: "a"},
	) != 0 {
		t.Fatalf("expected equal standings to compare as zero")
	}
}

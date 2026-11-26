package dynamodb_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

// Missing offers surface the domain sentinel so callers can map 404.
func TestOfferStore_Get_MissingReturnsErrOfferNotAvailable(t *testing.T) {
	// Arrange
	store := newOfferStoreForList(t, "hbk-missions-offer-get-1")

	// Act
	_, err := store.Get("offer-missing")

	// Assert
	require.ErrorIs(t, err, domain.ErrOfferNotAvailable)
}

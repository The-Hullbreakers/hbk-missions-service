//go:build integration

package http_test

import (
	"net/http/httptest"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
	missiondynamo "github.com/hbk/hbk-missions-service/internal/missions/infrastructure/dynamodb"
)

// languageReviewQueueServer seeds pending-validation contracts built from
// the accepted bilingual/legacy fixtures, plus one accepted contract to
// prove queue filtering.
func languageReviewQueueServer(t *testing.T, bilingual, legacy bool) (*httptest.Server, error) {
	return languageHTTPServer(t, func(client *missiondynamo.DynamoDBClient, table string) error {
		store := missiondynamo.NewContractStore(client, table)
		if bilingual {
			pending, err := pendingContractFixture(true)
			if err != nil {
				return err
			}
			if err := store.Save(pending); err != nil {
				return err
			}
		}
		if legacy {
			pending, err := pendingContractFixture(false)
			if err != nil {
				return err
			}
			if err := store.Save(pending); err != nil {
				return err
			}
			accepted := domain.NewMissionContract("contract-accepted", pending.Offer, []string{"player-1"})
			accepted.Terms = pending.Terms
			if err := store.Save(accepted); err != nil {
				return err
			}
		}
		return nil
	})
}

// pendingContractFixture rebuilds a fixture contract in PENDING_VALIDATION.
func pendingContractFixture(bilingual bool) (*domain.MissionContract, error) {
	var (
		source *domain.MissionContract
		err    error
	)
	if bilingual {
		source, err = bilingualContractFixture([]string{"player-1"})
	} else {
		source, err = legacyContractFixture([]string{"player-1"})
	}
	if err != nil {
		return nil, err
	}
	pending := domain.NewMissionContractFromState(source.ContractID, source.Offer,
		source.Roster(), domain.ContractPendingValidation, true, "",
		source.SeasonID, nil, nil, nil)
	pending.Terms = source.Terms
	return pending, nil
}

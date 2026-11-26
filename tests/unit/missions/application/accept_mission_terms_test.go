package application_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/hbk/hbk-missions-service/internal/missions/application"
	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

type termsOfferStore struct{ offer domain.MissionOffer }

func (s termsOfferStore) Get(string) (domain.MissionOffer, error) { return s.offer, nil }

type noActiveContracts struct{}

func (noActiveContracts) ListActiveByPlayer(string) ([]domain.MissionContract, error) {
	return nil, nil
}

type historyContracts struct{ contracts []domain.MissionContract }

func (s historyContracts) ListByPlayer(string, domain.ContractListFilter) ([]domain.MissionContract, error) {
	return s.contracts, nil
}

type savingContractWriter struct{ saved *domain.MissionContract }

func (w *savingContractWriter) CommitAcceptance(c *domain.MissionContract, _ domain.MissionOffer, _ int) error {
	w.saved = c
	return nil
}

type versionedDefinitions struct{ definition domain.MissionDefinition }

func (s versionedDefinitions) Get(string) (domain.MissionDefinition, error) {
	return s.definition, nil
}

func (s versionedDefinitions) GetVersion(string, string) (domain.MissionDefinition, error) {
	return s.definition, nil
}

func acceptedOfferFixture() domain.MissionOffer {
	offer := domain.NewMissionOffer("off-1", "BOUNTY", "season-1")
	offer.DefinitionCode = "def-1"
	offer.DefinitionVersion = "v1"
	return offer
}

func TestAcceptMission_FreezesDefinitionTerms(t *testing.T) {
	definition := domain.MissionDefinition{
		DefinitionCode: "def-1", Category: "BOUNTY",
		Mode: domain.MissionModeSolo, State: domain.DefinitionStatePublished, Version: "v1",
		Objectives: []domain.Objective{domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 5, 3)},
	}
	writer := &savingContractWriter{}
	handler := application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(3, nil),
		termsOfferStore{offer: acceptedOfferFixture()},
		noActiveContracts{},
		historyContracts{},
		versionedDefinitions{definition: definition},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		nil,
	)
	identity, _ := domain.NewCallerIdentity("player-1", nil, nil)

	contract, err := handler.Handle(identity, "off-1", "", nil)

	require.NoError(t, err)
	objective, ok := contract.Terms.Lookup("obj-a")
	require.True(t, ok, "expected contract terms to carry def-1 objectives")
	require.False(t, objective.MeritPerUnit != 10 || objective.Bonus != 5 || objective.UnitLimit != 3)
}

func TestAcceptMission_ReAcceptGeneratesUniqueAttempt(t *testing.T) {
	prior := lifecycleFixture(t, nil)
	prior.ContractID = "off-1:player-1:1"
	prior.State = domain.ContractAbandoned
	writer := &savingContractWriter{}
	handler := application.NewAcceptMissionHandler(
		domain.NewActiveLimitPolicy(3, nil),
		termsOfferStore{offer: acceptedOfferFixture()},
		noActiveContracts{},
		historyContracts{contracts: []domain.MissionContract{*prior}},
		versionedDefinitions{definition: domain.MissionDefinition{DefinitionCode: "def-1", Category: "BOUNTY", Mode: domain.MissionModeSolo, State: domain.DefinitionStatePublished, Version: "v1", Objectives: []domain.Objective{domain.NewObjectiveNamed("obj-a", "obj-a", "", "BOUNTY", 10, 0, 0)}}},
		fixedSeasonStore{season: domain.NewSeason("season-1", domain.SeasonStateActive)},
		writer,
		nil,
	)
	identity, _ := domain.NewCallerIdentity("player-1", nil, nil)

	contract, err := handler.Handle(identity, "off-1", "", nil)

	require.NoError(t, err)
	require.False(t, contract.ContractID == "" || contract.ContractID == prior.ContractID)
}

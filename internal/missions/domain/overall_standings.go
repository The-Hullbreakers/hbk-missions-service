package domain

import "sort"

type overallStandingFacts struct {
	standing           Standing
	completedContracts map[string]struct{}
	objectivePairs     map[[2]string]struct{}
}

// ProjectOverallStandings ranks distinct individuals by merits earned during one season.
func ProjectOverallStandings(credits []MeritCredit, completed map[string]bool, seasonID string) []Standing {
	byRecipient := make(map[string]*overallStandingFacts)
	for _, credit := range credits {
		if credit.SeasonID != seasonID || credit.Track != TrackIndividual {
			continue
		}
		facts, ok := byRecipient[credit.RecipientID]
		if !ok {
			facts = &overallStandingFacts{
				standing:           Standing{SeasonID: seasonID, Track: TrackIndividual, RecipientID: credit.RecipientID},
				completedContracts: make(map[string]struct{}),
				objectivePairs:     make(map[[2]string]struct{}),
			}
			byRecipient[credit.RecipientID] = facts
		}
		facts.standing.Total += credit.Amount
		if completed[credit.ContractID] {
			facts.completedContracts[credit.ContractID] = struct{}{}
		}
		facts.objectivePairs[[2]string{credit.Category, credit.ObjectiveID}] = struct{}{}
	}
	standings := make([]Standing, 0, len(byRecipient))
	for _, facts := range byRecipient {
		facts.standing.CompletedMissions = len(facts.completedContracts)
		facts.standing.UniqueObjectives = len(facts.objectivePairs)
		standings = append(standings, facts.standing)
	}
	sort.Slice(standings, func(i, j int) bool {
		return CompareStanding(standings[i], standings[j]) < 0
	})
	return standings
}

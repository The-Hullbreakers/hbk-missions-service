package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hbk/hbk-missions-service/internal/missions/domain"
)

func seasonFromItem(item map[string]types.AttributeValue) domain.Season {
	official := true
	if v, ok := item["Official"].(*types.AttributeValueMemberBOOL); ok {
		official = v.Value
	}
	season := domain.NewSeasonWithOfficial(getString(item, "SK"), domain.SeasonState(getString(item, "State")), official)
	season.StartsAt = getTime(item, "StartsAt")
	season.Schedule = scheduleFromItem(item)
	return season
}

func scheduleFromItem(item map[string]types.AttributeValue) *domain.SeasonSchedule {
	prefix := getString(item, "ScheduleSeriesPrefix")
	iteration := getInt(item, "ScheduleSeriesIteration")
	activeWeeks := getInt(item, "ScheduleActiveWeeks")
	offWeeks := getInt(item, "ScheduleOffWeeks")
	if prefix == "" || iteration <= 0 || activeWeeks <= 0 {
		return nil
	}
	return &domain.SeasonSchedule{
		ActiveWeeks: activeWeeks,
		OffWeeks:    offWeeks,
		Series: domain.SeasonSeries{
			Prefix:    prefix,
			Iteration: iteration,
		},
	}
}

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
	activeDays := getInt(item, "ScheduleActiveDays")
	offDays := getInt(item, "ScheduleOffDays")
	activeWeeks := getInt(item, "ScheduleActiveWeeks")
	offWeeks := getInt(item, "ScheduleOffWeeks")
	// Legacy items written before CHG-0044 only carry week attributes.
	if activeDays == 0 && activeWeeks > 0 {
		activeDays = activeWeeks * 7
	}
	if offDays == 0 && offWeeks > 0 {
		offDays = offWeeks * 7
	}
	if prefix == "" || iteration < 0 || activeDays <= 0 {
		return nil
	}
	return &domain.SeasonSchedule{
		ActiveDays:  activeDays,
		OffDays:     offDays,
		ActiveWeeks: activeWeeks,
		OffWeeks:    offWeeks,
		Series: domain.SeasonSeries{
			Prefix:    prefix,
			Iteration: iteration,
		},
	}
}

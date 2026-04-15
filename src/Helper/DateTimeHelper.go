package Helper

import (
	"time"
)

func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func TimestampToFirstAndLastMinuteOfDay(timestamp int64) (time.Time, time.Time) {
	t := time.Unix(timestamp, 0)
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-time.Nanosecond)
	return startOfDay, endOfDay
}

func YearsBetweenDates(fromDate int64, toDate int64) []int {
	fromTime := TimestampToTime(fromDate)
	toTime := TimestampToTime(toDate)
	var years []int

	for currentYear := fromTime.Year(); currentYear <= toTime.Year(); currentYear++ {
		years = append(years, currentYear)
	}
	return years
}

func GetUpcomingDays(timeRange int) []time.Time {
	var days []time.Time
	today := time.Now()
	startDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	for i := 0; i < timeRange; i++ {
		monthStart := startDate.AddDate(0, i, 0)
		monthEnd := monthStart.AddDate(0, 1, -1)
		for d := monthStart; !d.After(monthEnd); d = d.AddDate(0, 0, 1) {
			days = append(days, d)
		}
	}
	return days
}

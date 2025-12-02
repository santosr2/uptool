// Copyright (c) 2024 santosr2
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Interval constants for schedules.
const (
	intervalDaily        = "daily"
	intervalWeekly       = "weekly"
	intervalMonthly      = "monthly"
	intervalQuarterly    = "quarterly"
	intervalSemiannually = "semiannually"
	intervalYearly       = "yearly"
)

// ScheduleChecker handles schedule validation and enforcement.
type ScheduleChecker struct {
	schedule *Schedule
	timezone *time.Location
}

// NewScheduleChecker creates a new schedule checker.
func NewScheduleChecker(schedule *Schedule) (*ScheduleChecker, error) {
	if schedule == nil {
		return &ScheduleChecker{}, nil
	}

	var tz *time.Location
	var err error

	if schedule.Timezone != "" {
		tz, err = time.LoadLocation(schedule.Timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", schedule.Timezone, err)
		}
	} else {
		tz = time.UTC
	}

	return &ScheduleChecker{
		schedule: schedule,
		timezone: tz,
	}, nil
}

// ShouldRun checks if an update should run based on the schedule.
// Returns true if no schedule is configured or if the current time matches the schedule.
func (sc *ScheduleChecker) ShouldRun(now time.Time) bool {
	if sc.schedule == nil || sc.schedule.Interval == "" {
		return true
	}

	// Convert to configured timezone
	now = now.In(sc.timezone)

	// Check time of day if specified
	if sc.schedule.Time != "" {
		if !sc.isWithinTimeWindow(now) {
			return false
		}
	}

	switch strings.ToLower(sc.schedule.Interval) {
	case intervalDaily:
		return true // If time matches (or no time specified), run daily

	case intervalWeekly:
		return sc.isScheduledDay(now)

	case intervalMonthly:
		// Run on the first day of the month (or configured day)
		return now.Day() == 1

	case intervalQuarterly:
		// Run on the first day of each quarter (Jan, Apr, Jul, Oct)
		return now.Day() == 1 && (now.Month() == 1 || now.Month() == 4 || now.Month() == 7 || now.Month() == 10)

	case intervalSemiannually:
		// Run twice a year (Jan and Jul)
		return now.Day() == 1 && (now.Month() == 1 || now.Month() == 7)

	case intervalYearly:
		// Run on January 1st
		return now.Day() == 1 && now.Month() == 1

	case "cron":
		return sc.matchesCron(now)

	default:
		// Unknown interval, allow by default
		return true
	}
}

// isWithinTimeWindow checks if the current time is within 1 hour of the scheduled time.
func (sc *ScheduleChecker) isWithinTimeWindow(now time.Time) bool {
	if sc.schedule.Time == "" {
		return true
	}

	parts := strings.Split(sc.schedule.Time, ":")
	if len(parts) != 2 {
		return true // Invalid format, allow
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return true
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return true
	}

	// Create scheduled time for today
	scheduled := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, sc.timezone)

	// Allow a 1-hour window around the scheduled time
	windowStart := scheduled.Add(-30 * time.Minute)
	windowEnd := scheduled.Add(30 * time.Minute)

	return now.After(windowStart) && now.Before(windowEnd)
}

// isScheduledDay checks if today is the configured day for weekly schedules.
func (sc *ScheduleChecker) isScheduledDay(now time.Time) bool {
	if sc.schedule.Day == "" {
		// Default to Monday if no day specified
		return now.Weekday() == time.Monday
	}

	day := strings.ToLower(sc.schedule.Day)
	weekday := now.Weekday()

	switch day {
	case "monday":
		return weekday == time.Monday
	case "tuesday":
		return weekday == time.Tuesday
	case "wednesday":
		return weekday == time.Wednesday
	case "thursday":
		return weekday == time.Thursday
	case "friday":
		return weekday == time.Friday
	case "saturday":
		return weekday == time.Saturday
	case "sunday":
		return weekday == time.Sunday
	default:
		return false
	}
}

// matchesCron checks if the current time matches the cron expression.
// Supports standard 5-field cron format: minute hour day-of-month month day-of-week
func (sc *ScheduleChecker) matchesCron(now time.Time) bool {
	if sc.schedule.Cron == "" {
		return true
	}

	parts := strings.Fields(sc.schedule.Cron)
	if len(parts) != 5 {
		// Invalid cron format, allow by default
		return true
	}

	// Parse cron fields
	minute := parts[0]
	hour := parts[1]
	dayOfMonth := parts[2]
	month := parts[3]
	dayOfWeek := parts[4]

	// Check each field
	if !matchCronField(minute, now.Minute()) {
		return false
	}
	if !matchCronField(hour, now.Hour()) {
		return false
	}
	if !matchCronField(dayOfMonth, now.Day()) {
		return false
	}
	if !matchCronField(month, int(now.Month())) {
		return false
	}
	if !matchCronField(dayOfWeek, int(now.Weekday())) {
		return false
	}

	return true
}

// matchCronField checks if a value matches a cron field pattern.
// Supports: *, specific numbers, ranges (1-5), steps (*/2), lists (1,3,5)
func matchCronField(field string, value int) bool {
	field = strings.TrimSpace(field)

	// Wildcard matches everything
	if field == "*" {
		return true
	}

	// Handle step values (*/n)
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(strings.TrimPrefix(field, "*/"))
		if err != nil || step <= 0 {
			return true // Invalid step, allow
		}
		return value%step == 0
	}

	// Handle lists (1,3,5)
	if strings.Contains(field, ",") {
		for _, part := range strings.Split(field, ",") {
			if matchCronField(part, value) {
				return true
			}
		}
		return false
	}

	// Handle ranges (1-5)
	if strings.Contains(field, "-") {
		rangeParts := strings.Split(field, "-")
		if len(rangeParts) != 2 {
			return true // Invalid range, allow
		}
		start, err1 := strconv.Atoi(rangeParts[0])
		end, err2 := strconv.Atoi(rangeParts[1])
		if err1 != nil || err2 != nil {
			return true // Invalid numbers, allow
		}
		return value >= start && value <= end
	}

	// Exact match
	exactVal, err := strconv.Atoi(field)
	if err != nil {
		return true // Invalid number, allow
	}
	return value == exactVal
}

// GetNextRunTime calculates the next scheduled run time.
func (sc *ScheduleChecker) GetNextRunTime(from time.Time) time.Time {
	if sc.schedule == nil || sc.schedule.Interval == "" {
		return from
	}

	from = from.In(sc.timezone)

	// Get scheduled time of day
	hour, minute := 0, 0
	if sc.schedule.Time != "" {
		parts := strings.Split(sc.schedule.Time, ":")
		if len(parts) == 2 {
			hour, _ = strconv.Atoi(parts[0])   //nolint:errcheck // default to 0 on parse error
			minute, _ = strconv.Atoi(parts[1]) //nolint:errcheck // default to 0 on parse error
		}
	}

	switch strings.ToLower(sc.schedule.Interval) {
	case intervalDaily:
		next := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, sc.timezone)
		if next.Before(from) || next.Equal(from) {
			next = next.AddDate(0, 0, 1)
		}
		return next

	case intervalWeekly:
		// Find next occurrence of the scheduled day
		targetDay := sc.parseWeekday(sc.schedule.Day)
		next := time.Date(from.Year(), from.Month(), from.Day(), hour, minute, 0, 0, sc.timezone)

		daysUntilTarget := (int(targetDay) - int(next.Weekday()) + 7) % 7
		if daysUntilTarget == 0 && (next.Before(from) || next.Equal(from)) {
			daysUntilTarget = 7
		}
		return next.AddDate(0, 0, daysUntilTarget)

	case intervalMonthly:
		next := time.Date(from.Year(), from.Month(), 1, hour, minute, 0, 0, sc.timezone)
		if next.Before(from) || next.Equal(from) {
			next = next.AddDate(0, 1, 0)
		}
		return next

	case intervalQuarterly:
		// Next quarter start
		currentMonth := from.Month()
		var nextQuarterMonth time.Month
		switch {
		case currentMonth < 4:
			nextQuarterMonth = 4
		case currentMonth < 7:
			nextQuarterMonth = 7
		case currentMonth < 10:
			nextQuarterMonth = 10
		default:
			nextQuarterMonth = 1
		}

		year := from.Year()
		if nextQuarterMonth == 1 {
			year++
		}
		return time.Date(year, nextQuarterMonth, 1, hour, minute, 0, 0, sc.timezone)

	case intervalSemiannually:
		currentMonth := from.Month()
		var nextMonth time.Month
		year := from.Year()
		if currentMonth < 7 {
			nextMonth = 7
		} else {
			nextMonth = 1
			year++
		}
		return time.Date(year, nextMonth, 1, hour, minute, 0, 0, sc.timezone)

	case intervalYearly:
		next := time.Date(from.Year(), 1, 1, hour, minute, 0, 0, sc.timezone)
		if next.Before(from) || next.Equal(from) {
			next = next.AddDate(1, 0, 0)
		}
		return next

	default:
		return from
	}
}

// parseWeekday converts a day name to time.Weekday.
func (sc *ScheduleChecker) parseWeekday(day string) time.Weekday {
	switch strings.ToLower(day) {
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	case "sunday":
		return time.Sunday
	default:
		return time.Monday // Default
	}
}

// GetScheduleDescription returns a human-readable schedule description.
func (sc *ScheduleChecker) GetScheduleDescription() string {
	if sc.schedule == nil || sc.schedule.Interval == "" {
		return "on demand"
	}

	var desc strings.Builder

	switch strings.ToLower(sc.schedule.Interval) {
	case intervalDaily:
		desc.WriteString(intervalDaily)
	case intervalWeekly:
		desc.WriteString(intervalWeekly)
		if sc.schedule.Day != "" {
			desc.WriteString(" on ")
			caser := cases.Title(language.English)
			desc.WriteString(caser.String(sc.schedule.Day))
		}
	case intervalMonthly:
		desc.WriteString(intervalMonthly)
	case intervalQuarterly:
		desc.WriteString(intervalQuarterly)
	case intervalSemiannually:
		desc.WriteString("twice yearly")
	case intervalYearly:
		desc.WriteString(intervalYearly)
	case "cron":
		desc.WriteString("cron: ")
		desc.WriteString(sc.schedule.Cron)
	default:
		desc.WriteString(sc.schedule.Interval)
	}

	if sc.schedule.Time != "" {
		desc.WriteString(" at ")
		desc.WriteString(sc.schedule.Time)
	}

	if sc.schedule.Timezone != "" && sc.schedule.Timezone != "UTC" {
		desc.WriteString(" ")
		desc.WriteString(sc.schedule.Timezone)
	}

	return desc.String()
}

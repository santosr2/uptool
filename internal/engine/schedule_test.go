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
	"testing"
	"time"
)

func TestNewScheduleChecker(t *testing.T) {
	tests := []struct {
		schedule *Schedule
		name     string
		wantErr  bool
	}{
		{
			name:     "nil schedule",
			schedule: nil,
			wantErr:  false,
		},
		{
			name:     "empty schedule",
			schedule: &Schedule{},
			wantErr:  false,
		},
		{
			name: "valid timezone",
			schedule: &Schedule{
				Interval: "weekly",
				Timezone: "America/New_York",
			},
			wantErr: false,
		},
		{
			name: "invalid timezone",
			schedule: &Schedule{
				Interval: "weekly",
				Timezone: "Invalid/Timezone",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewScheduleChecker(tt.schedule)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewScheduleChecker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScheduleChecker_ShouldRun_Daily(t *testing.T) {
	schedule := &Schedule{
		Interval: "daily",
	}
	checker, err := NewScheduleChecker(schedule)
	if err != nil {
		t.Fatalf("NewScheduleChecker() error = %v", err)
	}

	// Daily should always run
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	if !checker.ShouldRun(now) {
		t.Error("Daily schedule should always run")
	}
}

func TestScheduleChecker_ShouldRun_Weekly(t *testing.T) {
	tests := []struct {
		testDate time.Time
		name     string
		day      string
		want     bool
	}{
		{
			name:     "Monday matches",
			day:      "monday",
			testDate: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC), // Monday
			want:     true,
		},
		{
			name:     "Monday doesn't match Tuesday",
			day:      "monday",
			testDate: time.Date(2025, 1, 14, 10, 0, 0, 0, time.UTC), // Tuesday
			want:     false,
		},
		{
			name:     "Friday matches",
			day:      "friday",
			testDate: time.Date(2025, 1, 17, 10, 0, 0, 0, time.UTC), // Friday
			want:     true,
		},
		{
			name:     "Default to Monday",
			day:      "",
			testDate: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC), // Monday
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := &Schedule{
				Interval: "weekly",
				Day:      tt.day,
			}
			checker, err := NewScheduleChecker(schedule)
			if err != nil {
				t.Fatalf("NewScheduleChecker() error = %v", err)
			}

			got := checker.ShouldRun(tt.testDate)
			if got != tt.want {
				t.Errorf("ShouldRun() = %v, want %v (day: %s)", got, tt.want, tt.testDate.Weekday())
			}
		})
	}
}

func TestScheduleChecker_ShouldRun_Monthly(t *testing.T) {
	schedule := &Schedule{
		Interval: "monthly",
	}
	checker, err := NewScheduleChecker(schedule)
	if err != nil {
		t.Fatalf("NewScheduleChecker() error = %v", err)
	}

	// First of month
	firstOfMonth := time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC)
	if !checker.ShouldRun(firstOfMonth) {
		t.Error("Monthly schedule should run on first of month")
	}

	// Other day
	otherDay := time.Date(2025, 2, 15, 10, 0, 0, 0, time.UTC)
	if checker.ShouldRun(otherDay) {
		t.Error("Monthly schedule should not run on other days")
	}
}

func TestScheduleChecker_ShouldRun_Quarterly(t *testing.T) {
	schedule := &Schedule{
		Interval: "quarterly",
	}
	checker, err := NewScheduleChecker(schedule)
	if err != nil {
		t.Fatalf("NewScheduleChecker() error = %v", err)
	}

	tests := []struct {
		date time.Time
		name string
		want bool
	}{
		{name: "Jan 1", date: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), want: true},
		{name: "Apr 1", date: time.Date(2025, 4, 1, 10, 0, 0, 0, time.UTC), want: true},
		{name: "Jul 1", date: time.Date(2025, 7, 1, 10, 0, 0, 0, time.UTC), want: true},
		{name: "Oct 1", date: time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC), want: true},
		{name: "Feb 1", date: time.Date(2025, 2, 1, 10, 0, 0, 0, time.UTC), want: false},
		{name: "Jan 15", date: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.ShouldRun(tt.date)
			if got != tt.want {
				t.Errorf("ShouldRun(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestScheduleChecker_ShouldRun_Yearly(t *testing.T) {
	schedule := &Schedule{
		Interval: "yearly",
	}
	checker, err := NewScheduleChecker(schedule)
	if err != nil {
		t.Fatalf("NewScheduleChecker() error = %v", err)
	}

	// Jan 1
	jan1 := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)
	if !checker.ShouldRun(jan1) {
		t.Error("Yearly schedule should run on Jan 1")
	}

	// Other day
	otherDay := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
	if checker.ShouldRun(otherDay) {
		t.Error("Yearly schedule should not run on other days")
	}
}

func TestScheduleChecker_ShouldRun_WithTime(t *testing.T) {
	schedule := &Schedule{
		Interval: "daily",
		Time:     "09:00",
	}
	checker, err := NewScheduleChecker(schedule)
	if err != nil {
		t.Fatalf("NewScheduleChecker() error = %v", err)
	}

	tests := []struct {
		time time.Time
		name string
		want bool
	}{
		{name: "within window (exact)", time: time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC), want: true},
		{name: "within window (before)", time: time.Date(2025, 1, 15, 8, 35, 0, 0, time.UTC), want: true},
		{name: "within window (after)", time: time.Date(2025, 1, 15, 9, 25, 0, 0, time.UTC), want: true},
		{name: "outside window (too early)", time: time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC), want: false},
		{name: "outside window (too late)", time: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checker.ShouldRun(tt.time)
			if got != tt.want {
				t.Errorf("ShouldRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduleChecker_ShouldRun_Cron(t *testing.T) {
	tests := []struct {
		time time.Time
		name string
		cron string
		want bool
	}{
		{
			name: "every Monday at 9am",
			cron: "0 9 * * 1",
			time: time.Date(2025, 1, 13, 9, 0, 0, 0, time.UTC), // Monday 9:00
			want: true,
		},
		{
			name: "every Monday at 9am - wrong day",
			cron: "0 9 * * 1",
			time: time.Date(2025, 1, 14, 9, 0, 0, 0, time.UTC), // Tuesday 9:00
			want: false,
		},
		{
			name: "every day at noon",
			cron: "0 12 * * *",
			time: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "first of month",
			cron: "0 0 1 * *",
			time: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "every 15 minutes",
			cron: "*/15 * * * *",
			time: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "every 15 minutes - no match",
			cron: "*/15 * * * *",
			time: time.Date(2025, 1, 15, 10, 17, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "range - hour 9-17",
			cron: "0 9-17 * * *",
			time: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			want: true,
		},
		{
			name: "range - hour outside",
			cron: "0 9-17 * * *",
			time: time.Date(2025, 1, 15, 18, 0, 0, 0, time.UTC),
			want: false,
		},
		{
			name: "list - specific hours",
			cron: "0 9,12,15 * * *",
			time: time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := &Schedule{
				Interval: "cron",
				Cron:     tt.cron,
			}
			checker, err := NewScheduleChecker(schedule)
			if err != nil {
				t.Fatalf("NewScheduleChecker() error = %v", err)
			}

			got := checker.ShouldRun(tt.time)
			if got != tt.want {
				t.Errorf("ShouldRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchCronField(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value int
		want  bool
	}{
		{name: "wildcard", field: "*", value: 5, want: true},
		{name: "exact match", field: "5", value: 5, want: true},
		{name: "exact no match", field: "5", value: 6, want: false},
		{name: "step */5 match", field: "*/5", value: 10, want: true},
		{name: "step */5 no match", field: "*/5", value: 11, want: false},
		{name: "range match", field: "1-5", value: 3, want: true},
		{name: "range no match", field: "1-5", value: 6, want: false},
		{name: "list match", field: "1,3,5", value: 3, want: true},
		{name: "list no match", field: "1,3,5", value: 4, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchCronField(tt.field, tt.value)
			if got != tt.want {
				t.Errorf("matchCronField(%q, %d) = %v, want %v", tt.field, tt.value, got, tt.want)
			}
		})
	}
}

func TestScheduleChecker_GetNextRunTime(t *testing.T) {
	tests := []struct {
		from     time.Time
		want     time.Time
		schedule *Schedule
		name     string
	}{
		{
			name: "daily next day",
			schedule: &Schedule{
				Interval: "daily",
				Time:     "09:00",
			},
			from: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), // After 9am
			want: time.Date(2025, 1, 16, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "daily same day",
			schedule: &Schedule{
				Interval: "daily",
				Time:     "09:00",
			},
			from: time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC), // Before 9am
			want: time.Date(2025, 1, 15, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "weekly next occurrence",
			schedule: &Schedule{
				Interval: "weekly",
				Day:      "friday",
				Time:     "09:00",
			},
			from: time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC), // Monday
			want: time.Date(2025, 1, 17, 9, 0, 0, 0, time.UTC),  // Friday
		},
		{
			name: "monthly next month",
			schedule: &Schedule{
				Interval: "monthly",
				Time:     "09:00",
			},
			from: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC), // Jan 15
			want: time.Date(2025, 2, 1, 9, 0, 0, 0, time.UTC),   // Feb 1
		},
		{
			name: "yearly next year",
			schedule: &Schedule{
				Interval: "yearly",
				Time:     "09:00",
			},
			from: time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC), // June 15, 2025
			want: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),   // Jan 1, 2026
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := NewScheduleChecker(tt.schedule)
			if err != nil {
				t.Fatalf("NewScheduleChecker() error = %v", err)
			}

			got := checker.GetNextRunTime(tt.from)
			if !got.Equal(tt.want) {
				t.Errorf("GetNextRunTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScheduleChecker_GetScheduleDescription(t *testing.T) {
	tests := []struct {
		name     string
		schedule *Schedule
		want     string
	}{
		{
			name:     "nil schedule",
			schedule: nil,
			want:     "on demand",
		},
		{
			name:     "empty interval",
			schedule: &Schedule{},
			want:     "on demand",
		},
		{
			name:     "daily",
			schedule: &Schedule{Interval: "daily"},
			want:     "daily",
		},
		{
			name:     "weekly with day",
			schedule: &Schedule{Interval: "weekly", Day: "monday"},
			want:     "weekly on Monday",
		},
		{
			name:     "daily with time",
			schedule: &Schedule{Interval: "daily", Time: "09:00"},
			want:     "daily at 09:00",
		},
		{
			name:     "weekly with timezone",
			schedule: &Schedule{Interval: "weekly", Day: "monday", Time: "09:00", Timezone: "America/New_York"},
			want:     "weekly on Monday at 09:00 America/New_York",
		},
		{
			name:     "cron",
			schedule: &Schedule{Interval: "cron", Cron: "0 9 * * 1"},
			want:     "cron: 0 9 * * 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker, err := NewScheduleChecker(tt.schedule)
			if err != nil {
				t.Fatalf("NewScheduleChecker() error = %v", err)
			}

			got := checker.GetScheduleDescription()
			if got != tt.want {
				t.Errorf("GetScheduleDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}

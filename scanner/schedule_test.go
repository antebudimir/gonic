package scanner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClock(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in      string
		expect  Clock
		wantErr bool
	}{
		{in: "04:30", expect: Clock{Hour: 4, Minute: 30}},
		{in: "00:00", expect: Clock{}},
		{in: "23:59", expect: Clock{Hour: 23, Minute: 59}},
		{in: "4:5", expect: Clock{Hour: 4, Minute: 5}},
		{in: "", wantErr: true},
		{in: "0430", wantErr: true},
		{in: "04:30:00", wantErr: true},
		{in: "24:00", wantErr: true},
		{in: "12:60", wantErr: true},
		{in: "-1:30", wantErr: true},
		{in: "aa:bb", wantErr: true},
	} {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			clock, err := ParseClock(tc.in)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expect, clock)
		})
	}
}

func TestClockString(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "04:30", Clock{Hour: 4, Minute: 30}.String())
	assert.Equal(t, "23:05", Clock{Hour: 23, Minute: 5}.String())
}

func TestScheduleDue(t *testing.T) {
	t.Parallel()
	at := Clock{Hour: 4, Minute: 30}
	day := func(hour, minute int) time.Time {
		return time.Date(2026, 10, 18, hour, minute, 0, 0, time.UTC)
	}

	for _, tc := range []struct {
		name   string
		now    time.Time
		last   time.Time
		days   int
		at     Clock
		expect bool
	}{
		{"disabled", day(5, 0), day(0, 0), 0, at, false},
		{"before time of day", day(4, 29), day(0, 0), 1, at, false},
		{"at time of day, never scanned", day(4, 30), time.Time{}, 1, at, true},
		{"after time of day, never scanned", day(9, 0), time.Time{}, 3, at, true},
		{"interval not elapsed", day(5, 0), day(5, 0), 1, at, false},
		{"interval elapsed exactly", day(5, 0).AddDate(0, 0, 1), day(5, 0), 1, at, true},
		{"interval elapsed, before time of day", day(4, 29).AddDate(0, 0, 1), day(5, 0), 1, at, false},
		{"multi day interval elapsed", day(6, 0).AddDate(0, 0, 3), day(6, 0), 3, at, true},
		{"multi day interval not elapsed", day(6, 0).AddDate(0, 0, 2), day(6, 0), 3, at, false},
		{"manual scan earlier today does not postpone", day(20, 0), day(19, 45), 1, at, false}, // before time of day
		{"scheduled scan earlier today blocks same slot", day(5, 0), day(5, 0), 1, at, false},
		{"time changed to later today fires today", day(20, 0), day(19, 45), 1, Clock{Hour: 20, Minute: 0}, true},
		{"manual scan never postpones (last is scheduled only)", day(4, 31).AddDate(0, 0, 1), time.Time{}, 1, at, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expect, scheduleDue(tc.now, tc.last, tc.days, tc.at))
		})
	}
}

func TestNextRun(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 18, 19, 0, 0, 0, time.UTC)
	at := Clock{Hour: 4, Minute: 30}

	t.Run("no scheduled scan, time today not reached yet", func(t *testing.T) {
		t.Parallel()
		got := NextRun(now, time.Time{}, 1, Clock{Hour: 20, Minute: 0})
		want := time.Date(2026, 10, 18, 20, 0, 0, 0, time.UTC)
		assert.Equal(t, want, got)
	})

	t.Run("no last scan, time later today", func(t *testing.T) {
		t.Parallel()
		got := NextRun(now, time.Time{}, 1, Clock{Hour: 20, Minute: 0})
		want := time.Date(2026, 10, 18, 20, 0, 0, 0, time.UTC)
		assert.Equal(t, want, got)
	})

	t.Run("last scheduled scan covered today slot, shows tomorrow", func(t *testing.T) {
		t.Parallel()
		last := time.Date(2026, 10, 18, 12, 0, 0, 0, time.UTC)
		got := NextRun(now, last, 1, at)
		want := time.Date(2026, 10, 19, 4, 30, 0, 0, time.UTC)
		assert.Equal(t, want, got)
	})

	t.Run("time changed to later today, shows today", func(t *testing.T) {
		t.Parallel()
		last := time.Date(2026, 10, 18, 19, 45, 0, 0, time.UTC)
		got := NextRun(now, last, 1, Clock{Hour: 20, Minute: 0})
		want := time.Date(2026, 10, 18, 20, 0, 0, 0, time.UTC)
		assert.Equal(t, want, got)
	})

	t.Run("no scheduled scan yet and slot already passed, due now", func(t *testing.T) {
		t.Parallel()
		got := NextRun(now, time.Time{}, 1, at)
		want := time.Date(2026, 10, 18, 4, 30, 0, 0, time.UTC)
		assert.Equal(t, want, got)
	})

	t.Run("disabled schedule", func(t *testing.T) {
		t.Parallel()
		assert.True(t, NextRun(now, time.Time{}, 0, at).IsZero())
	})
}

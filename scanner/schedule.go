package scanner

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"go.senan.xyz/gonic/db"
)

type Clock struct {
	Hour   int
	Minute int
}

func ParseClock(in string) (Clock, error) {
	hourStr, minuteStr, ok := strings.Cut(in, ":")
	if !ok {
		return Clock{}, fmt.Errorf("invalid time %q", in)
	}
	hour, err := strconv.Atoi(hourStr)
	if err != nil {
		return Clock{}, fmt.Errorf("invalid hour in %q", in)
	}
	minute, err := strconv.Atoi(minuteStr)
	if err != nil {
		return Clock{}, fmt.Errorf("invalid minute in %q", in)
	}
	if clock := (Clock{Hour: hour, Minute: minute}); !clock.valid() {
		return Clock{}, fmt.Errorf("time out of range in %q", in)
	}
	return Clock{Hour: hour, Minute: minute}, nil
}

func (c Clock) valid() bool {
	return c.Hour >= 0 && c.Hour < 24 && c.Minute >= 0 && c.Minute < 60
}

func (c Clock) Before(o Clock) bool {
	if c.Hour != o.Hour {
		return c.Hour < o.Hour
	}
	return c.Minute < o.Minute
}

func (c Clock) String() string {
	return fmt.Sprintf("%02d:%02d", c.Hour, c.Minute)
}

type ScanSchedule struct {
	Enabled bool
	Days    int
	Time    Clock
}

const scheduleDefaultDays = 1

func scheduleDefaultTime() Clock { return Clock{Hour: 4, Minute: 30} }

func (s *Scanner) Schedule() (*ScanSchedule, error) {
	enabledStr, err := s.db.GetSetting(db.ScanScheduleEnabled)
	if err != nil {
		return nil, fmt.Errorf("get schedule enabled: %w", err)
	}
	daysStr, err := s.db.GetSetting(db.ScanScheduleDays)
	if err != nil {
		return nil, fmt.Errorf("get schedule days: %w", err)
	}
	timeStr, err := s.db.GetSetting(db.ScanScheduleTime)
	if err != nil {
		return nil, fmt.Errorf("get schedule time: %w", err)
	}

	schedule := &ScanSchedule{Enabled: enabledStr == "1", Days: scheduleDefaultDays, Time: scheduleDefaultTime()}
	if daysStr != "" {
		days, err := strconv.Atoi(daysStr)
		if err != nil || days < 1 {
			return nil, fmt.Errorf("invalid schedule days %q", daysStr)
		}
		schedule.Days = days
	}
	if timeStr != "" {
		clock, err := ParseClock(timeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid schedule time: %w", err)
		}
		schedule.Time = clock
	}
	return schedule, nil
}

func (s *Scanner) SetSchedule(schedule *ScanSchedule) error {
	enabledStr := map[bool]string{true: "1", false: "0"}[schedule.Enabled]
	if err := s.db.SetSetting(db.ScanScheduleEnabled, enabledStr); err != nil {
		return fmt.Errorf("set schedule enabled: %w", err)
	}
	if err := s.db.SetSetting(db.ScanScheduleDays, strconv.Itoa(schedule.Days)); err != nil {
		return fmt.Errorf("set schedule days: %w", err)
	}
	if err := s.db.SetSetting(db.ScanScheduleTime, schedule.Time.String()); err != nil {
		return fmt.Errorf("set schedule time: %w", err)
	}
	return nil
}

// scheduleDue reports whether a scan is due. the schedule is anchored to the
// time-of-day it was configured with: a run fires the first time the clock
// reaches `at` on a day whose slot was not already covered by a previous
// scheduled scan. scans made outside the schedule (manual, subsonic, etc.)
// are never counted here, so they can not postpone a run, and changing the
// schedule time takes effect the same day if that time is still ahead.
func scheduleDue(now, lastScheduled time.Time, days int, at Clock) bool {
	if days < 1 {
		return false
	}
	nowClock := Clock{Hour: now.Hour(), Minute: now.Minute()}
	if nowClock.Before(at) {
		return false
	}
	if lastScheduled.IsZero() {
		return true
	}
	slot := time.Date(now.Year(), now.Month(), now.Day(), at.Hour, at.Minute, 0, 0, now.Location())
	return lastScheduled.Before(slot.AddDate(0, 0, -(days - 1)))
}

// NextRun returns the next time a scan will run on the given schedule, as of
// now, given that the last scheduled scan completed at lastScheduled. if the
// current slot is uncovered the scan is due immediately and the slot time is
// returned.
func NextRun(now, lastScheduled time.Time, days int, at Clock) time.Time {
	if days < 1 {
		return time.Time{}
	}
	candidate := time.Date(now.Year(), now.Month(), now.Day(), at.Hour, at.Minute, 0, 0, now.Location())
	for {
		if lastScheduled.IsZero() || lastScheduled.Before(candidate.AddDate(0, 0, -(days-1))) {
			return candidate
		}
		candidate = candidate.AddDate(0, 0, days)
	}
}

func (s *Scanner) LastScheduledScanTime() time.Time {
	tStr, err := s.db.GetSetting(db.LastScheduledScanTime)
	if err != nil || tStr == "" {
		return time.Time{}
	}
	i, _ := strconv.ParseInt(tStr, 10, 64)
	return time.Unix(i, 0)
}

func (s *Scanner) ExecuteSchedule(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		schedule, err := s.Schedule()
		if err != nil {
			log.Printf("error reading scan schedule: %v", err)
			continue
		}
		if !schedule.Enabled {
			continue
		}
		if !scheduleDue(time.Now(), s.LastScheduledScanTime(), schedule.Days, schedule.Time) {
			continue
		}
		if _, err := s.ScanAndClean(ScanOptions{}); err != nil && !errors.Is(err, ErrAlreadyScanning) {
			log.Printf("error scanning on schedule: %v", err)
			continue
		}
		if err := s.db.SetSetting(db.LastScheduledScanTime, strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
			log.Printf("error setting last scheduled scan time: %v", err)
		}
	}
}

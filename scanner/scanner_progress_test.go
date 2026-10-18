package scanner_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/mockfs"
	"go.senan.xyz/gonic/scanner"
)

func TestFoldersProcessed(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)

	require.Zero(t, m.Scanner().FoldersProcessed())

	m.AddItems()
	m.ScanAndClean()

	first := m.Scanner().FoldersProcessed()
	require.Positive(t, first)

	m.ScanAndClean()
	assert.Equal(t, first, m.Scanner().FoldersProcessed())
}

func TestFoldersProcessedWithExcludePattern(t *testing.T) {
	t.Parallel()
	m := mockfs.NewWithExcludePattern(t, `\/artist-1\/`)

	m.AddItems()
	m.ScanAndClean()

	excluded := m.Scanner().FoldersProcessed()

	unrestricted := mockfs.New(t)
	unrestricted.AddItems()
	unrestricted.ScanAndClean()

	assert.Less(t, excluded, unrestricted.Scanner().FoldersProcessed())
}

func TestScheduleRoundTrip(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)
	m.AddItems()
	s := m.Scanner()

	schedule, err := s.Schedule()
	require.NoError(t, err)
	assert.False(t, schedule.Enabled)
	assert.Equal(t, 1, schedule.Days)
	assert.Equal(t, scanner.Clock{Hour: 4, Minute: 30}, schedule.Time)

	schedule = &scanner.ScanSchedule{Enabled: true, Days: 3, Time: scanner.Clock{Hour: 22, Minute: 15}}
	require.NoError(t, s.SetSchedule(schedule))

	saved, err := s.Schedule()
	require.NoError(t, err)
	assert.Equal(t, schedule, saved)
}

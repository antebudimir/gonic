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

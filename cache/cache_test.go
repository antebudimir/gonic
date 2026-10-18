package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/cache"
)

func TestUsage(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := cache.New(dir, 0)

	writeCacheFile(t, dir, "a", 100)
	writeCacheFile(t, dir, "sub/b", 250)

	files, bytes, err := c.Usage()
	require.NoError(t, err)
	assert.Equal(t, 2, files)
	assert.Equal(t, int64(350), bytes)
}

func TestUsageEmptyCache(t *testing.T) {
	t.Parallel()
	c := cache.New(t.TempDir(), 0)

	files, bytes, err := c.Usage()
	require.NoError(t, err)
	assert.Zero(t, files)
	assert.Zero(t, bytes)
}

func TestEject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := cache.New(dir, 1)

	oldest := writeCacheFile(t, dir, "oldest", 600<<10)
	middle := writeCacheFile(t, dir, "middle", 600<<10)
	newest := writeCacheFile(t, dir, "newest", 100<<10)

	touch(t, oldest, time.Now().Add(-2*time.Hour))
	touch(t, middle, time.Now().Add(-1*time.Hour))
	touch(t, newest, time.Now())

	require.NoError(t, c.Eject())

	_, err := os.Stat(oldest)
	assert.Error(t, err, "oldest file should be ejected")

	_, err = os.Stat(middle)
	assert.NoError(t, err, "middle file should be kept, cache is within limit")
	_, err = os.Stat(newest)
	assert.NoError(t, err, "newest file should be kept")

	_, bytes, err := c.Usage()
	require.NoError(t, err)
	assert.LessOrEqual(t, bytes, int64(1)*1024*1024)
}

func TestSetLimitMB(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	c := cache.New(dir, 100)

	oldest := writeCacheFile(t, dir, "oldest", 900<<10)
	newest := writeCacheFile(t, dir, "newest", 300<<10)
	touch(t, oldest, time.Now().Add(-2*time.Hour))
	touch(t, newest, time.Now())

	// shrink the limit on a live cache, then eject
	c.SetLimitMB(1)
	assert.Equal(t, 1, c.LimitMB())

	require.NoError(t, c.Eject())

	_, err := os.Stat(oldest)
	assert.Error(t, err, "oldest file should be ejected under the new limit")
	_, err = os.Stat(newest)
	assert.NoError(t, err, "newest file should be kept")

	// raising to no limit keeps everything on the next eject
	c.SetLimitMB(0)
	assert.Equal(t, 0, c.LimitMB())
	require.NoError(t, c.Eject())
	_, err = os.Stat(newest)
	assert.NoError(t, err)
}

func writeCacheFile(t *testing.T, dir, name string, size time.Duration) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, make([]byte, int(size)), 0o644))
	return path
}

func touch(t *testing.T, path string, at time.Time) {
	t.Helper()
	require.NoError(t, os.Chtimes(path, at, at))
}

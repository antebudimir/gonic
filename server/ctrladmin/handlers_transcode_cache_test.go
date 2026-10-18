package ctrladmin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/cache"
	"go.senan.xyz/gonic/db"
	"go.senan.xyz/gonic/mockfs"
)

func TestServeUpdateTranscodeCacheSizeDo(t *testing.T) {
	t.Parallel()

	newController := func(t *testing.T) *Controller {
		m := mockfs.New(t)
		return &Controller{
			dbc:            m.DB(),
			scanner:        m.Scanner(),
			transcodeCache: cache.New(t.TempDir(), 10),
		}
	}
	post := func(t *testing.T, size string) *http.Request {
		t.Helper()
		r := httptest.NewRequest(http.MethodPost, "/admin/update_transcode_cache_size_do", strings.NewReader("size_mb="+size))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return r
	}

	t.Run("valid size is applied and persisted", func(t *testing.T) {
		c := newController(t)
		resp := c.ServeUpdateTranscodeCacheSizeDo(post(t, "250"))
		assert.Empty(t, resp.flashW)
		assert.Equal(t, 250, c.transcodeCache.LimitMB())
		v, err := c.dbc.GetSetting(db.TranscodeCacheSizeMB)
		require.NoError(t, err)
		assert.Equal(t, "250", v)
	})

	t.Run("zero means no limit", func(t *testing.T) {
		c := newController(t)
		resp := c.ServeUpdateTranscodeCacheSizeDo(post(t, "0"))
		assert.Empty(t, resp.flashW)
		assert.Equal(t, 0, c.transcodeCache.LimitMB())
		v, err := c.dbc.GetSetting(db.TranscodeCacheSizeMB)
		require.NoError(t, err)
		assert.Equal(t, "0", v)
	})

	t.Run("invalid size is rejected", func(t *testing.T) {
		c := newController(t)
		resp := c.ServeUpdateTranscodeCacheSizeDo(post(t, "abc"))
		assert.NotEmpty(t, resp.flashW)
		assert.Equal(t, 10, c.transcodeCache.LimitMB())
		resp = c.ServeUpdateTranscodeCacheSizeDo(post(t, "-5"))
		assert.NotEmpty(t, resp.flashW)
		assert.Equal(t, 10, c.transcodeCache.LimitMB())
	})
}

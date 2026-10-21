package ctrladmin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/db"
	"go.senan.xyz/gonic/mockfs"
	"go.senan.xyz/gonic/scanner"
)

func TestServeUpdateScanScheduleDo(t *testing.T) {
	t.Parallel()

	t.Run("saves valid schedule", func(t *testing.T) {
		t.Parallel()
		m := mockfs.New(t)
		c := &Controller{dbc: m.DB(), scanner: m.Scanner()}
		r := httptest.NewRequest(http.MethodPost, "/admin/update_scan_schedule_do", nil)
		require.NoError(t, r.ParseForm())
		r.Form.Set("enabled", "1")
		r.Form.Set("days", "3")
		r.Form.Set("time", "05:15")

		resp := c.ServeUpdateScanScheduleDo(r)
		assert.Empty(t, resp.flashW)

		schedule, err := m.Scanner().Schedule()
		require.NoError(t, err)
		assert.True(t, schedule.Enabled)
		assert.Equal(t, 3, schedule.Days)
		assert.Equal(t, scanner.Clock{Hour: 5, Minute: 15}, schedule.Time)
	})

	t.Run("rejects invalid days", func(t *testing.T) {
		t.Parallel()
		m := mockfs.New(t)
		c := &Controller{dbc: m.DB(), scanner: m.Scanner()}
		for _, days := range []string{"0", "-1", "31", "abc", ""} {
			r := httptest.NewRequest(http.MethodPost, "/admin/update_scan_schedule_do", nil)
			require.NoError(t, r.ParseForm())
			r.Form.Set("enabled", "1")
			r.Form.Set("days", days)
			r.Form.Set("time", "05:15")

			resp := c.ServeUpdateScanScheduleDo(r)
			assert.NotEmpty(t, resp.flashW, "days %q", days)
		}
	})

	t.Run("rejects invalid time", func(t *testing.T) {
		t.Parallel()
		m := mockfs.New(t)
		c := &Controller{dbc: m.DB(), scanner: m.Scanner()}
		for _, clock := range []string{"", "25:00", "nope"} {
			r := httptest.NewRequest(http.MethodPost, "/admin/update_scan_schedule_do", nil)
			require.NoError(t, r.ParseForm())
			r.Form.Set("enabled", "1")
			r.Form.Set("days", "1")
			r.Form.Set("time", clock)

			resp := c.ServeUpdateScanScheduleDo(r)
			assert.NotEmpty(t, resp.flashW, "time %q", clock)
		}
	})

	t.Run("can disable", func(t *testing.T) {
		t.Parallel()
		m := mockfs.New(t)
		c := &Controller{dbc: m.DB(), scanner: m.Scanner()}
		r := httptest.NewRequest(http.MethodPost, "/admin/update_scan_schedule_do", nil)
		require.NoError(t, r.ParseForm())
		r.Form.Set("enabled", "0")
		r.Form.Set("days", "2")
		r.Form.Set("time", "05:15")

		resp := c.ServeUpdateScanScheduleDo(r)
		assert.Empty(t, resp.flashW)

		schedule, err := m.Scanner().Schedule()
		require.NoError(t, err)
		assert.False(t, schedule.Enabled)
	})
}

func TestServeScanStatus(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)
	m.AddItems()
	m.ScanAndClean()

	c := &Controller{dbc: m.DB(), scanner: m.Scanner()}

	fetchStatus := func(t *testing.T) scanStatus {
		t.Helper()
		rr := httptest.NewRecorder()
		c.ServeScanStatus(rr, httptest.NewRequest(http.MethodGet, "/admin/scan_status", nil))
		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		var status scanStatus
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &status))
		return status
	}

	status := fetchStatus(t)
	assert.False(t, status.IsScanning)
	assert.Positive(t, status.FoldersProcessed)
	assert.NotZero(t, status.LastScanTime)

	require.True(t, c.scanner.StartScanning())
	defer c.scanner.StopScanning()

	status = fetchStatus(t)
	assert.True(t, status.IsScanning)

	lastScan := time.Now().Truncate(time.Second)
	require.NoError(t, m.DB().SetSetting(db.LastScanTime, strconv.FormatInt(lastScan.Unix(), 10)))

	status = fetchStatus(t)
	assert.Equal(t, lastScan.Unix(), status.LastScanTime)
}

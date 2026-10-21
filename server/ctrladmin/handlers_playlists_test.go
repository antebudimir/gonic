package ctrladmin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/db"
	"os"

	"go.senan.xyz/gonic/fileutil"
	"go.senan.xyz/gonic/mockfs"
	"go.senan.xyz/gonic/playlist"
)

func postRules(t *testing.T, c *Controller, form url.Values) *Response {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/admin/playlists", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.ServeUpdatePlaylistRulesDo(r)
}

func writeSmartPlaylist(t *testing.T, store *playlist.Store, relPath string, rules string) {
	t.Helper()
	require.NoError(t, store.Write(relPath, &playlist.Playlist{
		UserID:   1,
		Name:     "smart-jazz",
		Comment:  "do not touch",
		IsPublic: true,
		Rules:    rules,
		Items:    []string{"/music/a.flac", "/music/b.flac"},
	}))
}

// rawPlaylist reads the m3u bytes so byte-identical checks skip the parser.
func rawPlaylist(t *testing.T, store *playlist.Store, relPath string) string {
	t.Helper()
	abs, err := fileutil.SafeJoin(store.BasePath(), relPath)
	require.NoError(t, err)
	buf, err := os.ReadFile(abs)
	require.NoError(t, err)
	return string(buf)
}

// TestUpdatePlaylistRulesDo is the rules-edit POST matrix: this handler is the
// only admin path that writes to a user's playlist file on disk.
func TestUpdatePlaylistRulesDo(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)
	store, err := playlist.NewStore(t.TempDir())
	require.NoError(t, err)
	c := &Controller{dbc: m.DB(), playlistStore: store}
	relPath := filepath.Join("1", "smart.m3u")

	// valid rules: written canonically, existing items and frontmatter preserved
	writeSmartPlaylist(t, store, relPath, "")
	resp := postRules(t, c, url.Values{"playlist": {relPath}, "rules": {`{"genre":"Jazz","limit":10}`}})
	require.Empty(t, resp.err)
	require.NotEmpty(t, resp.flashN)

	pl, err := store.Read(relPath)
	require.NoError(t, err)
	require.Equal(t, `{"genre":"Jazz","limit":10}`, pl.Rules)
	require.Equal(t, "smart-jazz", pl.Name)
	require.Equal(t, "do not touch", pl.Comment)
	require.True(t, pl.IsPublic)
	require.Len(t, pl.Items, 2)

	// invalid JSON: flash a warning and leave the file byte-identical
	before := rawPlaylist(t, store, relPath)
	resp = postRules(t, c, url.Values{"playlist": {relPath}, "rules": {`{"genre":`}})
	require.NotEmpty(t, resp.flashW)
	require.Contains(t, resp.flashW[0], "invalid rules")
	require.Equal(t, before, rawPlaylist(t, store, relPath))

	// out-of-range values are rejected the same way
	resp = postRules(t, c, url.Values{"playlist": {relPath}, "rules": {`{"minRating":99}`}})
	require.NotEmpty(t, resp.flashW)
	require.Equal(t, before, rawPlaylist(t, store, relPath))

	// empty textarea clears #GONIC-RULES but keeps everything else
	resp = postRules(t, c, url.Values{"playlist": {relPath}, "rules": {""}})
	require.NotEmpty(t, resp.flashN)
	pl, err = store.Read(relPath)
	require.NoError(t, err)
	require.Equal(t, "", pl.Rules)
	require.Equal(t, "smart-jazz", pl.Name)
	require.Len(t, pl.Items, 2)

	// missing playlist path is a warning, not a panic
	resp = postRules(t, c, url.Values{"rules": {"{}"}})
	require.Contains(t, resp.flashW[0], "no playlist given")

	// unknown playlist path is a warning
	resp = postRules(t, c, url.Values{"playlist": {"9/missing.m3u"}, "rules": {"{}"}})
	require.Contains(t, resp.flashW[0], "error reading playlist")
}

func TestServePlaylists(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)
	store, err := playlist.NewStore(t.TempDir())
	require.NoError(t, err)
	c := &Controller{dbc: m.DB(), playlistStore: store} // mockfs seeds the admin user (id 1)

	writeSmartPlaylist(t, store, filepath.Join("1", "smart.m3u"), `{"genre":"Jazz"}`)
	require.NoError(t, store.Write(filepath.Join("1", "static.m3u"), &playlist.Playlist{
		UserID: 1, Name: "static", Items: []string{"a.flac"},
	}))
	require.NoError(t, store.Write(filepath.Join("1", "broken.m3u"), &playlist.Playlist{
		UserID: 1, Name: "broken", Rules: `{"sort":"nope"}`,
	}))

	r := httptest.NewRequest(http.MethodGet, "/admin/playlists", nil)
	resp := c.ServePlaylists(r)
	require.Empty(t, resp.err)
	views := resp.data.AllPlaylists
	require.Len(t, views, 3)

	byName := map[string]*playlistView{}
	for _, v := range views {
		byName[v.Name] = v
	}
	require.Contains(t, byName["smart-jazz"].RulesJSON, "Jazz")
	require.False(t, byName["static"].RulesInvalid)
	require.Equal(t, "", byName["static"].RulesJSON)
	require.Equal(t, "admin", byName["static"].Owner)
	require.True(t, byName["broken"].RulesInvalid)
	require.Contains(t, byName["broken"].RulesJSON, `"sort":"nope"`)
}

func TestServeShares(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)
	dbc := m.DB()
	var admin db.User // mockfs seeds the admin user
	require.NoError(t, dbc.Where("name=?", "admin").First(&admin).Error)

	expiry := time.Now().Add(-time.Hour)
	share := &db.Share{
		Token: "abcdefgh123456", UserID: admin.ID, Description: "d",
		VisitCount: 3, Downloadable: true, ExpiresAt: &expiry,
		Entries: []*db.ShareEntry{{Item: "tr-1"}, {Item: "tr-2"}},
	}
	require.NoError(t, dbc.Create(share).Error)

	c := &Controller{dbc: dbc}
	r := httptest.NewRequest(http.MethodGet, "/admin/shares", nil)
	r.Host = "gonic.example"
	resp := c.ServeShares(r)
	require.Empty(t, resp.err)
	require.Len(t, resp.data.AllShares, 1)

	views := map[string]*shareView{}
	for _, v := range resp.data.AllShares {
		views[v.Token] = v
	}
	v := views["abcdefgh\u2026"]
	require.NotNil(t, v)
	require.Equal(t, "http://gonic.example/share/abcdefgh123456", v.URL)
	require.Equal(t, "admin", v.Owner)
	require.Equal(t, 2, v.Entries)
	require.Equal(t, 3, v.VisitCount)
	require.True(t, v.Downloadable)
	require.True(t, v.Expired)
}

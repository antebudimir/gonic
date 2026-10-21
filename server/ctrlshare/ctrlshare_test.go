package ctrlshare

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go.senan.xyz/gonic/db"
	_ "go.senan.xyz/gonic/deps"
	"go.senan.xyz/gonic/server/ctrlsubsonic"
	paramsp "go.senan.xyz/gonic/server/ctrlsubsonic/params"
	"go.senan.xyz/gonic/server/ctrlsubsonic/specid"
)

func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	dbc, err := db.NewMock(url.Values{})
	if err != nil {
		t.Fatalf("new mock db: %v", err)
	}
	if err := dbc.AutoMigrate(
		db.User{}, db.Share{}, db.ShareEntry{}, db.Album{}, db.Track{},
	).Error; err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return dbc
}

func seedShare(t *testing.T, dbc *db.DB) (*db.User, *db.Share, *db.Album, *db.Track) {
	t.Helper()
	user := &db.User{Name: "admin", Password: "pw", IsAdmin: true}
	if err := dbc.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	album := &db.Album{RootDir: "/m", LeftPath: "artist-a", RightPath: "album-a", TagTitle: "album-a"}
	if err := dbc.Create(album).Error; err != nil {
		t.Fatalf("create album: %v", err)
	}
	track := &db.Track{AlbumID: album.ID, Filename: "track-0.flac", TagTitle: "track-a"}
	if err := dbc.Create(track).Error; err != nil {
		t.Fatalf("create track: %v", err)
	}
	share := &db.Share{
		Token: "testtoken", UserID: user.ID, Description: "my share",
		Entries: []*db.ShareEntry{{Item: track.SID().String()}},
	}
	if err := dbc.Create(share).Error; err != nil {
		t.Fatalf("create share: %v", err)
	}
	return user, share, album, track
}

func TestLandingAndGuards(t *testing.T) {
	t.Parallel()
	dbc := newTestDB(t)
	user, share, album, track := seedShare(t, dbc)
	c := New(dbc, nil) // subsonic delegation is covered by handler tests in ctrlsubsonic

	// landing page
	r := httptest.NewRequest(http.MethodGet, "/testtoken", nil)
	r.URL.Path = "/testtoken"
	w := httptest.NewRecorder()
	c.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("landing status %d body %q", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "my share") || !strings.Contains(w.Body.String(), "track-a") {
		t.Fatalf("landing missing content: %s", w.Body.String())
	}
	var after db.Share
	if err := dbc.Where("token=?", "testtoken").First(&after).Error; err != nil {
		t.Fatalf("reload share: %v", err)
	}
	if after.VisitCount != 1 || after.LastVisitedAt == nil {
		t.Fatalf("visit not counted: %d %v", after.VisitCount, after.LastVisitedAt)
	}

	// an album share expands into its tracks on the landing page
	albumShare := &db.Share{
		Token: "albumtoken", UserID: user.ID, Downloadable: true,
		Entries: []*db.ShareEntry{{Item: album.SID().String()}},
	}
	if err := dbc.Create(albumShare).Error; err != nil {
		t.Fatalf("create album share: %v", err)
	}
	r = httptest.NewRequest(http.MethodGet, "/albumtoken", nil)
	r.URL.Path = "/albumtoken"
	w = httptest.NewRecorder()
	c.ServeHTTP(w, r)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "id=tr-"+fmt.Sprint(track.ID)) {
		t.Fatalf("album landing should link track tr-%d, got: %s", track.ID, w.Body.String())
	}

	// stream of a foreign entry is forbidden
	r = httptest.NewRequest(http.MethodGet, "/testtoken/stream?id=tr-999", nil)
	r.URL.Path = "/testtoken/stream"
	r.URL.RawQuery = "id=tr-999"
	w = httptest.NewRecorder()
	c.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("foreign entry status %d body %q", w.Code, w.Body.String())
	}

	// downloadable=false blocks downloads before delegation
	r = httptest.NewRequest(http.MethodGet, "/testtoken/download?id="+track.SID().String(), nil)
	r.URL.Path = "/testtoken/download"
	r.URL.RawQuery = "id=" + track.SID().String()
	w = httptest.NewRecorder()
	c.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Fatalf("non-downloadable status %d body %q", w.Code, w.Body.String())
	}

	// expired share is gone
	expires := time.Now().Add(-time.Hour)
	if err := dbc.Model(share).Update("expires_at", expires).Error; err != nil {
		t.Fatalf("expire share: %v", err)
	}
	r = httptest.NewRequest(http.MethodGet, "/testtoken", nil)
	r.URL.Path = "/testtoken"
	w = httptest.NewRecorder()
	c.ServeHTTP(w, r)
	if w.Code != http.StatusGone {
		t.Fatalf("expired status %d body %q", w.Code, w.Body.String())
	}
}

func TestDownloadDisposition(t *testing.T) {
	dbc := newTestDB(t)
	_, _, _, track := seedShare(t, dbc)
	w := httptest.NewRecorder()
	setDownloadDisposition(w, dbc, *track.SID())
	cd := w.Header().Get("Content-Disposition")
	if !strings.HasPrefix(cd, "attachment;") || !strings.Contains(cd, "track-0.flac") {
		t.Fatalf("expected attachment filename track-0.flac, got %q", cd)
	}
	// unknown track: no header, delegation proceeds without disposition
	w = httptest.NewRecorder()
	setDownloadDisposition(w, dbc, specid.ID{Type: specid.Track, Value: 999})
	if w.Header().Get("Content-Disposition") != "" {
		t.Fatalf("expected no header for unknown track")
	}
}

func TestEntryAllowed(t *testing.T) {
	t.Parallel()
	dbc := newTestDB(t)
	_, _, _, track := seedShare(t, dbc)

	entries := []specid.ID{*track.SID()}
	if !entryAllowed(dbc, entries, *track.SID()) {
		t.Fatal("exact entry should be allowed")
	}
	foreign := specid.ID{Type: specid.Track, Value: track.ID + 1}
	if entryAllowed(dbc, entries, foreign) {
		t.Fatal("foreign track should not be allowed")
	}
}

// ensure the synthetic context contract expected by ServeStream stays intact
func TestSyntheticContext(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodGet, "/x/stream?id=tr-1", nil)
	r.URL.RawQuery = "id=tr-1"
	user := &db.User{Name: "admin"}
	r = r.Clone(context.WithValue(r.Context(), ctrlsubsonic.CtxUser, user))
	params := paramsp.New(r)
	if _, err := params.GetID("id"); err != nil {
		t.Fatalf("params should see the query: %v", err)
	}
}

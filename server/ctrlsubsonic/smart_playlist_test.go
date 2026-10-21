package ctrlsubsonic

import (
	"testing"

	"go.senan.xyz/gonic/db"
	playlistp "go.senan.xyz/gonic/playlist"
)

// genreRuleCount resolves a smart playlist rule set and returns the track titles,
// so per-field tests can assert exact matches instead of golden files.
func genreRuleCount(t *testing.T, f *fixture, userID int, rules *playlistp.Rules) []string {
	t.Helper()
	rows, err := smartPlaylistTrackRows(f.contr, userID, rules)
	if err != nil {
		t.Fatalf("smart rows %v: %v", rules, err)
	}
	var titles []string
	for _, row := range rows {
		titles = append(titles, row.TagTitle)
	}
	return titles
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestSmartPlaylistEmptyRules(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	titles := genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{})
	if len(titles) == 0 {
		t.Fatal("empty rules should match all tracks, got none")
	}
}

func TestSmartPlaylistGenreTrackAndAlbum(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	// track-level genre: every Jazz match must come from a track_genres row
	titles := genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{Genre: "Jazz"})
	if len(titles) == 0 {
		t.Fatal("jazz: want at least one track")
	}
	for _, title := range titles {
		var n int64
		f.dbc.Model(&db.TrackGenre{}).
			Joins("JOIN genres g ON g.id = track_genres.genre_id").
			Joins("JOIN tracks t ON t.id = track_genres.track_id").
			Where("g.name=? AND t.tag_title=?", "Jazz", title).
			Count(&n)
		if n == 0 {
			t.Fatalf("jazz match %q has no track-level Jazz genre", title)
		}
	}

	// album-inheritance branch: genre reachable ONLY via album_genres.
	// the album join must match tracks that have no matching track_genres row.
	genre := &db.Genre{Name: "Blues"}
	if err := f.dbc.Create(genre).Error; err != nil {
		t.Fatalf("create genre: %v", err)
	}
	ag := &db.AlbumGenre{AlbumID: f.albumAA.ID, GenreID: genre.ID}
	if err := f.dbc.Create(ag).Error; err != nil {
		t.Fatalf("create album genre: %v", err)
	}
	titles = genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{Genre: "Blues"})
	if len(titles) != 3 {
		t.Fatalf("album-inherited genre: want all 3 album-aa tracks, got %v", titles)
	}
}

func TestSmartPlaylistRatingAndPlays(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	// fixture: trackAB1 is rated 3/5/5 by admin/alt/alt2 (avg 4.33)
	titles := genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{MinRating: 3})
	if len(titles) != 1 || titles[0] != "rich-title-d1" {
		t.Fatalf("minRating 3: want only rich-title-d1, got %v", titles)
	}
	titles = genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{MinRating: 4})
	if len(titles) != 1 || titles[0] != "rich-title-d1" {
		t.Fatalf("minRating 4: avg 4.33 should still match, got %v", titles)
	}
	// the filter is any-user semantics: alt rated 5, so even minRating 5 matches
	titles = genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{MinRating: 5})
	if len(titles) != 1 || titles[0] != "rich-title-d1" {
		t.Fatalf("minRating 5: any-user rating of 5 should match, got %v", titles)
	}
	// an unrated user sees the same track: the filter is not per-caller
	titles = genreRuleCount(t, f, f.alt.ID, &playlistp.Rules{MinRating: 4})
	if len(titles) != 1 || titles[0] != "rich-title-d1" {
		t.Fatalf("minRating is cross-user: want rich-title-d1, got %v", titles)
	}

	// fixture: album-aa tracks each played 7/3 times, album-ab tracks 1 time
	titles = genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{MinPlays: 2})
	if len(titles) != 4 || !contains(titles, "collab-track") || !contains(titles, "title-0") {
		t.Fatalf("minPlays 2: want album-aa tracks plus collab-track, got %v", titles)
	}
	// collab-track was played exactly 3 times; 7/3-count tracks fall below
	titles = genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{MinPlays: 3})
	if len(titles) != 1 || titles[0] != "collab-track" {
		t.Fatalf("minPlays 3: want only collab-track, got %v", titles)
	}
	titles = genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{MinPlays: 4})
	if len(titles) != 0 {
		t.Fatalf("minPlays 4 must exclude count-3 track, got %v", titles)
	}
}

func TestSmartPlaylistSortAndLimit(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	titles := genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{Sort: "title", Limit: 1})
	if len(titles) != 1 {
		t.Fatalf("limit 1: want 1 track, got %v", titles)
	}
	first := titles[0]
	all := genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{Sort: "title"})
	if len(all) == 0 || all[0] != first {
		t.Fatalf("sort=title with limit must return the globally first title; first=%q all[0]=%q", first, func() string {
			if len(all) > 0 {
				return all[0]
			}
			return ""
		}())
	}

	years := genreRuleCount(t, f, f.admin.ID, &playlistp.Rules{Sort: "year"})
	if len(years) < 2 {
		t.Fatalf("sort=year: want multiple tracks, got %v", years)
	}

	// unknown field must be ignored, not error
	titles = genreRuleCount(t, f, f.admin.ID, func() *playlistp.Rules {
		r, err := playlistp.ParseRules(`{"genre":"Jazz","wat":true}`)
		if err != nil {
			t.Fatalf("unknown field: %v", err)
		}
		return r
	}())
	if len(titles) != 3 {
		t.Fatalf("unknown field tolerated, want the 3 jazz tracks, got %v", titles)
	}
}

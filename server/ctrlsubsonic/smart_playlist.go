package ctrlsubsonic

import (
	playlistp "go.senan.xyz/gonic/playlist"
	"go.senan.xyz/gonic/server/ctrlsubsonic/spec"
)

// smartPlaylistTrackRows resolves a playlist's rules into tracks.
// all present rule fields are ANDed together. genre matching includes
// genres inherited from the track's album.
func smartPlaylistTrackRows(c *Controller, userID int, rules *playlistp.Rules) ([]*spec.TrackRow, error) {
	q := c.dbc.Scopes(spec.LoadTrackByTags(userID))

	if rules.Genre != "" {
		q = q.Where(
			"tracks.id IN (SELECT tg.track_id FROM track_genres tg JOIN genres g ON g.id=tg.genre_id WHERE g.name=?) OR "+
				"tracks.album_id IN (SELECT ag.album_id FROM album_genres ag JOIN genres g ON g.id=ag.genre_id WHERE g.name=?)",
			rules.Genre, rules.Genre)
	}
	if rules.FromYear > 0 {
		q = q.Where("tracks.tag_year>=?", rules.FromYear)
	}
	if rules.ToYear > 0 {
		q = q.Where("tracks.tag_year<=?", rules.ToYear)
	}
	if rules.MinRating > 0 {
		q = q.Where("tracks.id IN (SELECT track_id FROM track_ratings WHERE rating>=?)", rules.MinRating)
	}
	if rules.Starred {
		q = q.Where("tracks.id IN (SELECT track_id FROM track_stars)")
	}
	if rules.MinPlays > 0 {
		q = q.Where("tracks.id IN (SELECT track_id FROM track_plays WHERE count>=?)", rules.MinPlays)
	}

	switch rules.Sort {
	case "rating":
		q = q.Order("average_rating DESC")
	case "plays":
		q = q.Order("(SELECT coalesce(sum(count), 0) FROM track_plays tp WHERE tp.track_id=tracks.id) DESC")
	case "year":
		q = q.Order("tracks.tag_year")
	case "title":
		q = q.Order("tracks.tag_title")
	default: // "" and "random"
		q = q.Order("RANDOM()")
	}
	if rules.Limit > 0 {
		q = q.Limit(rules.Limit)
	}

	var rows []*spec.TrackRow
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

package ctrladmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"go.senan.xyz/gonic/playlist"
)

// playlistView is a read-only view of a playlist file for the admin UI.
type playlistView struct {
	// RelPath is the playlist path relative to the playlists dir, eg "1/smart-jazz.m3u".
	// it's shown so the user knows which file to edit to change the playlist.
	RelPath string
	Name    string
	Owner   string
	// RulesJSON is the pretty-printed #GONIC-RULES value, empty for static playlists.
	RulesJSON string
	// RulesInvalid is set when a smart playlist's rules fail to parse.
	RulesInvalid bool
	Items        int
}

func (c *Controller) ServePlaylists(r *http.Request) *Response {
	relPaths, err := c.playlistStore.List()
	if err != nil {
		return &Response{code: 500, err: fmt.Sprintf("error listing playlists: %v", err)}
	}
	sort.Strings(relPaths)

	var playlists []*playlistView
	for _, relPath := range relPaths {
		pl, err := c.playlistStore.Read(relPath)
		if err != nil {
			return &Response{code: 500, err: fmt.Sprintf("error reading playlist %q: %v", relPath, err)}
		}
		view := &playlistView{
			RelPath: relPath,
			Name:    pl.Name,
			Owner:   c.ownerName(pl.UserID),
			Items:   len(pl.Items),
		}
		if pl.Rules != "" {
			rules, err := playlist.ParseRules(pl.Rules)
			switch {
			case err != nil:
				view.RulesInvalid = true
				view.RulesJSON = pl.Rules
			default:
				buf, err := json.MarshalIndent(rules, "", "  ")
				if err != nil {
					view.RulesInvalid = true
					view.RulesJSON = pl.Rules
				} else {
					view.RulesJSON = string(buf)
				}
			}
		}
		playlists = append(playlists, view)
	}
	return &Response{template: "playlists.tmpl", data: &templateData{AllPlaylists: playlists}}
}

func (c *Controller) ServeUpdatePlaylistRulesDo(r *http.Request) *Response {
	const redirectPlaylists = "/admin/playlists"

	relPath := r.PostFormValue("playlist")
	if relPath == "" {
		return &Response{redirect: redirectPlaylists, flashW: []string{"no playlist given"}}
	}

	// validate before touching the file; empty input clears the rules
	rules := r.PostFormValue("rules")
	parsed, err := playlist.ParseRules(rules)
	if err != nil {
		return &Response{redirect: redirectPlaylists, flashW: []string{fmt.Sprintf("invalid rules: %v", err)}}
	}

	pl, err := c.playlistStore.Read(relPath)
	if err != nil {
		return &Response{redirect: redirectPlaylists, flashW: []string{fmt.Sprintf("error reading playlist: %v", err)}}
	}
	if parsed == nil {
		pl.Rules = ""
	} else {
		pl.Rules = parsed.String()
	}
	if err := c.playlistStore.Write(relPath, pl); err != nil {
		return &Response{redirect: redirectPlaylists, flashW: []string{fmt.Sprintf("error writing playlist: %v", err)}}
	}
	return &Response{redirect: redirectPlaylists, flashN: []string{"playlist rules updated"}}
}

func (c *Controller) ownerName(userID int) string {
	user, err := c.dbc.GetUserByID(userID)
	if err != nil || user == nil {
		return fmt.Sprintf("user %d", userID)
	}
	return user.Name
}

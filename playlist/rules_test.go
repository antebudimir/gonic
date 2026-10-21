package playlist_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.senan.xyz/gonic/playlist"
)

func TestRulesRoundTrip(t *testing.T) {
	t.Parallel()
	rules := &playlist.Rules{
		Genre: "Jazz", FromYear: 1960, ToYear: 1979,
		MinRating: 4, Starred: true, MinPlays: 5,
		Sort: "random", Limit: 50,
	}
	encoded := rules.String()

	store, err := playlist.NewStore(t.TempDir())
	require.NoError(t, err)
	require.NoError(t, store.Write("1", &playlist.Playlist{
		UserID: 1, Name: "smart", Rules: encoded,
		Items: []string{"/music/a.flac"},
	}))

	read, err := store.Read("1")
	require.NoError(t, err)
	require.Equal(t, encoded, read.Rules)

	parsed, err := playlist.ParseRules(read.Rules)
	require.NoError(t, err)
	require.Equal(t, rules, parsed)
}

func TestParseRules(t *testing.T) {
	t.Parallel()
	parsed, err := playlist.ParseRules("")
	require.NoError(t, err)
	require.Nil(t, parsed)

	_, err = playlist.ParseRules("{bad json")
	require.Error(t, err)

	_, err = playlist.ParseRules(`{"sort":"nope"}`)
	require.Error(t, err)

	_, err = playlist.ParseRules(`{"minRating":9}`)
	require.Error(t, err)

	_, err = playlist.ParseRules(`{"minRating":-1}`)
	require.Error(t, err)

	// whitespace-only is treated as no rules
	noRules, err := playlist.ParseRules("  \n\t ")
	require.NoError(t, err)
	require.Nil(t, noRules)

	// unknown fields are ignored, valid fields still parsed
	rules, err := playlist.ParseRules(`{"genre":"Jazz","fromFuture":3000}`)
	require.NoError(t, err)
	require.Equal(t, &playlist.Rules{Genre: "Jazz"}, rules)
}

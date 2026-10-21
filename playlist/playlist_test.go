package playlist_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.senan.xyz/gonic/playlist"
)

func TestPlaylist(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	store, err := playlist.NewStore(tmp)
	require.NoError(t, err)

	playlistIDs, err := store.List()
	require.NoError(t, err)
	require.Empty(t, playlistIDs)

	for _, playlistID := range playlistIDs {
		playlist, err := store.Read(playlistID)
		require.NoError(t, err)
		require.NotZero(t, playlist.UpdatedAt)
	}

	before := playlist.Playlist{
		UserID: 10,
		Name:   "Examlpe playlist name",
		Comment: `
Example comment
It has multiple lines 👍
`,
		Items: []string{
			"/music/item 1.flac",
			"/music/item 2.flac",
			"/music/item 3.flac",
		},
		IsPublic: true,
	}

	newPath := playlist.NewPath(before.UserID, before.Name)
	require.NoError(t, store.Write(newPath, &before))

	after, err := store.Read(newPath)
	require.NoError(t, err)

	require.Equal(t, after.UserID, before.UserID)
	require.Equal(t, after.Name, before.Name)
	require.Equal(t, after.Comment, before.Comment)
	require.Equal(t, after.Items, before.Items)
	require.Equal(t, after.IsPublic, before.IsPublic)

	playlistIDs, err = store.List()
	require.NoError(t, err)
	require.True(t, len(playlistIDs) == 1)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	store, err := playlist.NewStore(t.TempDir())
	require.NoError(t, err)

	require.NoError(t, store.Write("1/smart.m3u", &playlist.Playlist{UserID: 1, Name: "smart"}))
	require.NoError(t, store.Delete("1/smart.m3u"))

	ids, err := store.List()
	require.NoError(t, err)
	require.Empty(t, ids)

	// deleting again is an error, not a silent success
	require.Error(t, store.Delete("1/smart.m3u"))

	// traversal out of the base dir must be refused
	require.Error(t, store.Delete(filepath.Join("..", "evil.m3u")))
}

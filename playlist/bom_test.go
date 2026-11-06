package playlist_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/playlist"
)

// TestReadStripsUTF8BOM guards the fork's #GONIC-RULES attribute parsing: an m3u
// written with a leading UTF-8 BOM must still have its first attribute line
// (which is #GONIC-NAME) and its #GONIC-RULES attribute read correctly.
func TestReadStripsUTF8BOM(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	store, err := playlist.NewStore(tmp)
	require.NoError(t, err)

	want := playlist.Playlist{
		UserID: 1,
		Name:   "bom playlist",
		Rules:  "genre=rock",
	}
	relPath := playlist.NewPath(want.UserID, want.Name)
	require.NoError(t, store.Write(relPath, &want))

	// prepend a UTF-8 BOM, as written by some windows editors
	absPath := filepath.Join(tmp, relPath)
	body, err := os.ReadFile(absPath) //nolint:gosec // absPath is under t.TempDir()
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(absPath, append([]byte("\ufeff"), body...), 0o600)) //nolint:gosec // absPath is under t.TempDir()

	got, err := store.Read(relPath)
	require.NoError(t, err)
	require.Equal(t, want.Name, got.Name)
	require.Equal(t, want.Rules, got.Rules)
}

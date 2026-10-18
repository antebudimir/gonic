package podcast_test

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/podcast"
)

//go:embed testdata/opml_example.opml
var opmlExample []byte

func TestParseOPML(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		in      string
		expect  []string
		wantErr bool
	}{
		{
			name: "nested and loose outlines",
			in:   string(opmlExample),
			expect: []string{
				"https://example.com/news.xml",
				"https://example.com/news-2.xml",
				"https://example.com/comedy.xml",
				"https://example.com/loose.xml",
			},
		},
		{
			name:   "no feeds",
			in:     `<opml><body><outline text="grouping"><outline text="empty"/></outline></body></opml>`,
			expect: nil,
		},
		{
			name:   "empty body",
			in:     `<opml><body></body></opml>`,
			expect: nil,
		},
		{
			name:    "not xml",
			in:      "this is not opml",
			wantErr: true,
		},
		{
			name:    "wrong root element",
			in:      `<rss><channel></channel></rss>`,
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			urls, err := podcast.ParseOPML(strings.NewReader(tc.in))
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expect, urls)
		})
	}
}

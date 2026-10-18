package ctrladmin

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.senan.xyz/gonic/mockfs"
	"go.senan.xyz/gonic/podcast"
)

func TestServePodcastImportOPMLDo(t *testing.T) {
	t.Parallel()

	newController := func(t *testing.T) *Controller {
		t.Helper()
		m := mockfs.New(t)
		return &Controller{dbc: m.DB(), podcasts: podcast.New(m.DB(), t.TempDir(), m.TagReader(), "test"), scanner: m.Scanner()}
	}

	postOPML := func(t *testing.T, content []byte) *http.Request {
		t.Helper()
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		if content != nil {
			part, err := writer.CreateFormFile("opml", "podcasts.opml")
			assert.NoError(t, err)
			_, err = part.Write(content)
			assert.NoError(t, err)
		}
		assert.NoError(t, writer.Close())
		r := httptest.NewRequest(http.MethodPost, "/admin/import_podcasts_opml_do", body)
		r.Header.Set("Content-Type", writer.FormDataContentType())
		return r
	}

	t.Run("missing file warns", func(t *testing.T) {
		t.Parallel()
		c := newController(t)
		resp := c.ServePodcastImportOPMLDo(postOPML(t, nil))
		assert.NotEmpty(t, resp.flashW)
	})

	t.Run("invalid xml warns", func(t *testing.T) {
		t.Parallel()
		c := newController(t)
		resp := c.ServePodcastImportOPMLDo(postOPML(t, []byte("not opml")))
		assert.NotEmpty(t, resp.flashW)
	})

	t.Run("no feeds warns", func(t *testing.T) {
		t.Parallel()
		c := newController(t)
		resp := c.ServePodcastImportOPMLDo(postOPML(t, []byte(`<opml><body></body></opml>`)))
		assert.NotEmpty(t, resp.flashW)
	})
}

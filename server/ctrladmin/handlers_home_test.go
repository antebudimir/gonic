package ctrladmin

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"context"

	"github.com/Masterminds/sprig"
	"github.com/philippta/go-template/html/template"
	"github.com/stretchr/testify/require"

	"go.senan.xyz/gonic/db"
	"go.senan.xyz/gonic/mockfs"
	"go.senan.xyz/gonic/server/ctrladmin/adminui"
)

func TestServeHomeScheduleTimeRender(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)
	c := &Controller{dbc: m.DB(), scanner: m.Scanner()}

	r := httptest.NewRequest(http.MethodGet, "/admin/home", nil)
	r = r.WithContext(context.WithValue(r.Context(), CtxUser, &db.User{IsAdmin: true}))
	resp := c.ServeHome(r)
	if resp.err != "" {
		t.Fatalf("servehome err: %s", resp.err)
	}

	tmpl := template.Must(template.
		New("layout").
		Funcs(template.FuncMap(sprig.FuncMap())).
		Funcs(funcMap()).
		Funcs(template.FuncMap{"path": func(s string) string { return s }}).
		ParseFS(adminui.TemplatesFS, "*.tmpl", "**/*.tmpl"))

	resp.data.User = &db.User{IsAdmin: true, Name: "admin"}
	out := &strings.Builder{}
	if err := tmpl.ExecuteTemplate(out, resp.template, resp.data); err != nil {
		t.Fatalf("exec: %v", err)
	}
	re := regexp.MustCompile(`name="time"[^>]*`)
	line := re.FindString(out.String())
	t.Logf("RENDERED: %s", line)
	if !strings.Contains(line, "04:30") {
		t.Logf("OK: renders 04:30")
	}
	if strings.Contains(line, "map[") {
		t.Errorf("BUG REPRODUCED: %s", line)
	}
}

func TestServeHomePollingScriptEmbedsLastScanTime(t *testing.T) {
	t.Parallel()
	m := mockfs.New(t)
	c := &Controller{dbc: m.DB(), scanner: m.Scanner()}

	tmpl := template.Must(template.
		New("layout").
		Funcs(template.FuncMap(sprig.FuncMap())).
		Funcs(funcMap()).
		Funcs(template.FuncMap{"path": func(s string) string { return s }}).
		ParseFS(adminui.TemplatesFS, "*.tmpl", "**/*.tmpl"))

	render := func() string {
		r := httptest.NewRequest(http.MethodGet, "/admin/home", nil)
		r = r.WithContext(context.WithValue(r.Context(), CtxUser, &db.User{IsAdmin: true}))
		resp := c.ServeHome(r)
		if resp.err != "" {
			t.Fatalf("servehome err: %s", resp.err)
		}
		resp.data.User = &db.User{IsAdmin: true, Name: "admin"}
		out := &strings.Builder{}
		if err := tmpl.ExecuteTemplate(out, resp.template, resp.data); err != nil {
			t.Fatalf("exec: %v", err)
		}
		return out.String()
	}

	// no scan yet: initial last scan time embedded as 0
	out := render()
	condensed := strings.ReplaceAll(out, " ", "")
	if !strings.Contains(condensed, "constinitialLastScan=0;") {
		t.Errorf("expected initialLastScan = 0 with no scans, got page without it")
	}

	// after a scan: embedded as the scan's unix time
	last := time.Unix(1750000000, 0)
	require.NoError(t, m.DB().SetSetting(db.LastScanTime, strconv.FormatInt(last.Unix(), 10)))
	out = render()
	condensed = strings.ReplaceAll(out, " ", "")
	if !strings.Contains(condensed, "constinitialLastScan=1750000000;") {
		t.Errorf("expected initialLastScan = 1750000000 after a scan")
	}
}

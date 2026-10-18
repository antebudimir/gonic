package ctrladmin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"

	"go.senan.xyz/gonic/db"
)

func (c *Controller) ServeLoginDo(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(CtxSession).(*sessions.Session)
	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		sessAddFlashW(session, []string{"please provide username and password"})
		sessLogSave(session, w, r)
		http.Redirect(w, r, c.resolveProxyPath("/admin/login"), http.StatusSeeOther)
		return
	}
	user, err := c.dbc.GetUserByName(username)
	if err != nil {
		sessAddFlashW(session, []string{fmt.Sprintf("error finding user: %v", err)})
		sessLogSave(session, w, r)
		http.Redirect(w, r, c.resolveProxyPath("/admin/login"), http.StatusSeeOther)
		return
	}
	if user == nil || password != user.Password {
		sessAddFlashW(session, []string{"invalid username / password"})
		sessLogSave(session, w, r)
		http.Redirect(w, r, c.resolveProxyPath("/admin/login"), http.StatusSeeOther)
		return
	}
	// put the user name into the session. future endpoints after this one
	// are wrapped with WithUserSession() which will get the name from the
	// session and put the row into the request context
	session.Values["user"] = user.ID
	sessLogSave(session, w, r)
	http.Redirect(w, r, c.resolveProxyPath("/admin/home"), http.StatusSeeOther)
}

type scanStatus struct {
	IsScanning       bool  `json:"isScanning"`
	FoldersProcessed int   `json:"foldersProcessed"`
	LastScanTime     int64 `json:"lastScanTime,omitempty"`
}

func (c *Controller) ServeScanStatus(w http.ResponseWriter, _ *http.Request) {
	status := scanStatus{
		IsScanning:       c.scanner.IsScanning(),
		FoldersProcessed: c.scanner.FoldersProcessed(),
	}
	if tStr, err := c.dbc.GetSetting(db.LastScanTime); err != nil {
		http.Error(w, fmt.Sprintf("error getting last scan time: %v", err), 500)
		return
	} else if tStr != "" {
		i, _ := strconv.ParseInt(tStr, 10, 64)
		status.LastScanTime = i
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("error encoding scan status: %v\n", err)
	}
}

func (c *Controller) ServeLogout(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(CtxSession).(*sessions.Session)
	session.Options.MaxAge = -1
	sessLogSave(session, w, r)
	http.Redirect(w, r, c.resolveProxyPath("/admin/login"), http.StatusSeeOther)
}

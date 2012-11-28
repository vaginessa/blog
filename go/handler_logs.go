package main

import (
	"net/http"
)

// /logs
func handleLogs(w http.ResponseWriter, r *http.Request) {
	cookie := getSecureCookie(r)
	isAdmin := cookie.TwitterUser == "kjk" // only I can see the logs
	model := struct {
		UserIsAdmin bool
		Errors      []*TimestampedMsg
		Notices     []*TimestampedMsg
		Header      *http.Header
	}{
		UserIsAdmin: isAdmin,
	}

	if model.UserIsAdmin {
		model.Errors = logger.GetErrors()
		model.Notices = logger.GetNotices()
	}

	if r.FormValue("show") != "" {
		model.Header = &r.Header
		model.Header.Add("RealIp", getIpAddress(r))
	}

	ExecTemplate(w, tmplLogs, model)
}

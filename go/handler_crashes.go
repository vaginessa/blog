package main

import (
	"net/http"
)

// url: /app/crashshow?crash_id=${crash_id}
func handleCrashShow(w http.ResponseWriter, r *http.Request) {
	if !IsAdmin(r) {
		serve404(w, r)
		return
	}
	serve404(w, r)
}

func showCrashesIndex(w http.ResponseWriter, r *http.Request) {
	apps := storeCrashes.GetAppCrashInfos()
	model := struct {
		Apps []*AppCrashInfo
	}{
		Apps: apps,
	}
	ExecTemplate(w, tmplCrashReportsIndex, model)
}

// url: /app/crashes[?app_name=${app_name}]
func handleCrashes(w http.ResponseWriter, r *http.Request) {
	if !IsAdmin(r) {
		serve404(w, r)
		return
	}
	appName := getTrimmedFormValue(r, "app_name")
	if appName == "" {
		showCrashesIndex(w, r)
		return
	}

	serve404(w, r)
}

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

type CrashDisplay struct {
	Crash
	ShortCrashingLine string
}

// TODO: write me
func (c *CrashDisplay) CreatedOnSince() string {
	return ""
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
	crashes := storeCrashes.GetCrashesForApp(appName)
	n := len(crashes)
	dispCrashes := make([]CrashDisplay, n, n)
	for i, c := range crashes {
		dispCrashes[i] = CrashDisplay{Crash: *c}
	}
	model := struct {
		AppName string
		Crashes []CrashDisplay
	}{
		AppName: appName,
		Crashes: dispCrashes,
	}
	ExecTemplate(w, tmplCrashReportsAppIndex, model)
}

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

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

func readCrashReport(sha1 []byte) ([]byte, error) {
	return ReadFileAll(storeCrashes.MessageFilePath(sha1))
}

// url: /app/crashshow?crash_id=${crash_id}
func handleCrashShow(w http.ResponseWriter, r *http.Request) {
	if !IsAdmin(r) {
		serve404(w, r)
		return
	}
	crashIdStr := getTrimmedFormValue(r, "crash_id")
	crashId, err := strconv.Atoi(crashIdStr)
	if err != nil {
		serve404(w, r)
		return
	}
	crash := storeCrashes.GetCrashById(crashId)
	if crash == nil {
		serve404(w, r)
		return
	}
	appName := crash.AppCrashInfo.Name
	crashData, err := readCrashReport(crash.Sha1[:])
	if err != nil {
		serve404(w, r)
		return
	}

	crashBody := string(crashData)
	model := struct {
		IndexUrl  string
		IpAddr    string
		AppName   string
		CrashBody template.HTML
	}{
		IndexUrl:  fmt.Sprintf("/app/crashes?app_name=%s", appName),
		IpAddr:    crash.IpAddress(),
		AppName:   appName,
		CrashBody: template.HTML(crashBody),
	}
	ExecTemplate(w, tmplCrashReport, model)
}

package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (c *Crash) Version() string {
	ver := *c.ProgramVersion
	if ver == "" {
		return "no ver"
	}
	if strings.HasSuffix(ver, " pre-release") {
		return ver[:len(ver)-1-len(" pre-release")]
	}
	return ver
}

func TimeSinceNowAsString(t time.Time) string {
	d := time.Now().Sub(t)
	minutes := int(d.Minutes()) % 60
	hours := int(d.Hours())
	days := hours / 24
	hours = hours % 24
	if days > 0 {
		return fmt.Sprintf("%dd %dhr", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dhr %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func (c *Crash) CreatedOnSince() string {
	return TimeSinceNowAsString(c.CreatedOn)
}

type CrashesForDay struct {
	Day     string
	Crashes []*Crash
}

func (c *CrashesForDay) CrashesCount() int {
	return len(c.Crashes)
}

type AppDisplay struct {
	*App
	Days []CrashesForDay
}

// Reverse embeds a sort.Interface value and implements a reverse sort over
// that value.
type Reverse struct {
	// This embedded Interface permits Reverse to use the methods of
	// another Interface implementation.
	sort.Interface
}

// Less returns the opposite of the embedded implementation's Less method.
func (r Reverse) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}

func NewAppDisplay(app *App, addCrashesPerDay bool) *AppDisplay {
	res := &AppDisplay{App: app}
	if !addCrashesPerDay {
		res.Days = make([]CrashesForDay, 0)
		return res
	}

	n := len(app.PerDayCrashes)
	res.Days = make([]CrashesForDay, n, n)

	days := make([]string, n)
	i := 0
	for day, _ := range app.PerDayCrashes {
		days[i] = day
		i += 1
	}
	sort.Sort(Reverse{sort.StringSlice(days)})
	for i, day := range days {
		crashesForDay := CrashesForDay{Day: day}
		crashesForDay.Crashes = app.PerDayCrashes[day]
		res.Days[i] = crashesForDay
	}
	return res
}

func showCrashesIndex(w http.ResponseWriter, r *http.Request) {
	apps := storeCrashes.GetApps()
	model := struct {
		Apps []*App
	}{
		Apps: apps,
	}
	ExecTemplate(w, tmplCrashReportsIndex, model)
}

func showCrashesByIp(w http.ResponseWriter, r *http.Request, app *App, ipAddrInternal string) {
	appDisplay := NewAppDisplay(app, false)
	crashes := storeCrashes.GetCrashesForIpAddrInternal(app, ipAddrInternal)
	model := struct {
		App         *AppDisplay
		Crashes     []*Crash
		DayOrIpAddr string
	}{
		App:         appDisplay,
		Crashes:     crashes,
		DayOrIpAddr: crashes[0].IpAddress(),
	}
	ExecTemplate(w, tmplCrashReportsAppIndex, model)
}

// url: /app/crashes[?app_name=${appName}][&day=${day}][&ip_addr=${ipAddrInternal}]
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
	app := storeCrashes.GetAppByName(appName)
	if app == nil {
		logger.Errorf("handleCrashes(): invalid app '%s'", appName)
		serve404(w, r)
		return
	}

	ipAddrInternal := getTrimmedFormValue(r, "ip_addr")
	if ipAddrInternal != "" {
		showCrashesByIp(w, r, app, ipAddrInternal)
		return
	}

	day := getTrimmedFormValue(r, "day")

	appDisplay := NewAppDisplay(app, true)
	var crashes []*Crash
	for _, forDay := range appDisplay.Days {
		if day == forDay.Day {
			crashes = forDay.Crashes
			break
		}
	}
	if crashes == nil {
		crashes = appDisplay.Days[0].Crashes
		day = appDisplay.Days[0].Day
	}
	model := struct {
		App         *AppDisplay
		Crashes     []*Crash
		DayOrIpAddr string
	}{
		App:         appDisplay,
		Crashes:     crashes,
		DayOrIpAddr: day,
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
	crashData, err := readCrashReport(crash.Sha1[:])
	if err != nil {
		serve404(w, r)
		return
	}
	appName := crash.App.Name
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

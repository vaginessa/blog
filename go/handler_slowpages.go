package main

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	MaxPagesPerDay = 100
)

type PageTiming struct {
	Url            string
	GenerationTime time.Duration
	Timestamp      time.Time
}

type PageTimings struct {
	Timings        []*PageTiming
	FastestsOfSlow time.Duration
	Max            int
	Day            int
}

func (t *PageTiming) DurationStr() string {
	return fmt.Sprintf("%.4f secs", t.GenerationTime.Seconds())
}

var pageTimingsMutex sync.Mutex
var pageTimings = NewPageTimings(MaxPagesPerDay)

type PageTimingsByGenerationTime []*PageTiming

func (s PageTimingsByGenerationTime) Len() int {
	return len(s)
}
func (s PageTimingsByGenerationTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s PageTimingsByGenerationTime) Less(i, j int) bool {
	t1 := s[i].GenerationTime
	t2 := s[j].GenerationTime
	return t1.Nanoseconds() > t2.Nanoseconds()
}

func NewPageTimings(max int) *PageTimings {
	return &PageTimings{
		Timings:        make([]*PageTiming, 0),
		Max:            max,
		FastestsOfSlow: 0,
		Day:            time.Now().Day(),
	}
}

func (t *PageTimings) Add(url string, generationTime time.Duration) {
	if len(t.Timings) < t.Max {
		el := &PageTiming{url, generationTime, time.Now()}
		t.Timings = append(t.Timings, el)
		if generationTime > t.FastestsOfSlow {
			t.FastestsOfSlow = generationTime
		}
		if len(t.Timings) == t.Max {
			sort.Sort(PageTimingsByGenerationTime(t.Timings))
		}
		return
	}
	if generationTime < t.FastestsOfSlow {
		return
	}
	n := len(t.Timings) - 1
	t.Timings[n] = &PageTiming{url, generationTime, time.Now()}
	sort.Sort(PageTimingsByGenerationTime(t.Timings))
	t.FastestsOfSlow = t.Timings[n].GenerationTime
}

// make a copy to avoid multi-threading issues
func (t *PageTimings) GetTimings() []*PageTiming {
	n := len(t.Timings)
	res := make([]*PageTiming, n, n)
	for i, el := range t.Timings {
		res[i] = el
	}
	if n < t.Max {
		sort.Sort(PageTimingsByGenerationTime(res))
	}
	return res
}

func LogSlowPage(url string, generationTime time.Duration) {
	pageTimingsMutex.Lock()
	defer pageTimingsMutex.Unlock()
	day := time.Now().Day()
	if pageTimings.Day != day {
		pageTimings = NewPageTimings(MaxPagesPerDay)
	}
	pageTimings.Add(url, generationTime)
}

// /timings
func handleTimings(w http.ResponseWriter, r *http.Request) {
	cookie := getSecureCookie(r)
	isAdmin := cookie.TwitterUser == "kjk" // only I can see the logs
	pageTimingsMutex.Lock()
	timings := pageTimings.GetTimings()
	pageTimingsMutex.Unlock()

	isAdmin = true
	model := struct {
		UserIsAdmin bool
		PageTimings []*PageTiming
	}{
		UserIsAdmin: isAdmin,
	}

	if model.UserIsAdmin {
		model.PageTimings = timings
	}

	ExecTemplate(w, tmplTimings, model)
}

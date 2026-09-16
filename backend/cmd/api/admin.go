package main

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

type LogEntry struct {
	Path     string        `json:"path"`
	Duration time.Duration `json:"duration"`
}

type Stats struct {
	Endpoint string  `json:"endpoint"`
	Count    int     `json:"count"`
	MeanMS   float64 `json:"mean_ms"`
	MaxMS    float64 `json:"max_ms"`
}

// HANDLERS

func (app *app) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *app) requestsStatsHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("data/halendar_log.log")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer file.Close()

	stats := make(map[string]*Stats)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Ignore lines that don't look like JSON (slog might log non-JSON if misconfigured)
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		var path string
		var duration time.Duration

		if err := json.Unmarshal(raw["path"], &path); err != nil {
			continue
		}
		if err := json.Unmarshal(raw["duration"], &duration); err != nil {
			continue
		}

		stat := stats[path]
		if stat == nil {
			stat = &Stats{}
			stat.Endpoint = path
			stats[path] = stat
		}
		ms := float64(duration.Milliseconds())
		stat.Count++
		stat.MeanMS += ms
		if ms > stat.MaxMS {
			stat.MaxMS = ms
		}
	}

	// Finalize
	statsList := []*Stats{}
	for _, s := range stats {
		s.MeanMS /= float64(s.Count)
		statsList = append(statsList, s)
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"stats": statsList}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *app) backgroudProcess() {

	app.background(func() {
		for {
			now := time.Now()

			// Calculate the next "round" time (e.g., 12:05, 12:10, etc.)
			nextTime := now.Truncate(5 * time.Minute).Add(5 * time.Minute)
			if nextTime.Before(now) {
				nextTime = nextTime.Add(5 * time.Minute)
			}

			duration := nextTime.Sub(now)
			time.Sleep(duration)

			if app.config.env == "development" {
			} else {
				app.wg.Add(1)
				go app.cleanClientIPs()
			}
		}
	})

	app.background(func() {
		for {
			now := time.Now()

			// Calculate the next run time
			nextRun := calculateNextRun(now, []RunTime{{Hour: 23, Minute: 00}, {Hour: 06, Minute: 00}})
			duration := time.Until(nextRun)
			time.Sleep(duration)
		}
	})
}

func (app *app) cleanClientIPs() {
	defer app.wg.Done()

	app.mu.Lock()
	for ip, client := range app.clientsIPs {
		if time.Since(client.lastSeen) > 10*time.Minute {
			delete(app.clientsIPs, ip)
		}
	}
	app.mu.Unlock()
}

package calendar_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"halendar/calendar"
	"halendar/testutil"
)

var paris, _ = time.LoadLocation("Europe/Paris")

func tomorrow(hour int) time.Time {
	d := time.Now().In(paris).AddDate(0, 0, 1)
	return time.Date(d.Year(), d.Month(), d.Day(), hour, 0, 0, 0, paris)
}

func TestCalendarFullFlow(t *testing.T) {
	ctx := context.Background()
	f := testutil.NewFakeCalDAV(t)
	cal := calendar.New(calendar.Config{URL: f.URL, User: f.User, Pass: f.Pass})
	if err := cal.Test(ctx); err != nil {
		t.Fatal(err)
	}
	if names, _ := cal.Calendars(ctx); len(names) != 1 || names[0] != "Work" {
		t.Fatalf("calendars: %v", names)
	}

	// Add from JSON (local date + duration)
	var e calendar.Event
	raw := `{"title":"Hackathon pitch","start":"` + tomorrow(9).Format("2006-01-02T15:04") + `","duration_minutes":30,"status":"tentative","location":"Room 1","url":"https://call.example.com/x"}`
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		t.Fatal(err)
	}
	created, err := cal.Add(ctx, e)
	if err != nil || created.UID == "" || created.Calendar != "Work" || !created.End.Equal(tomorrow(9).Add(30*time.Minute)) {
		t.Fatalf("add: %+v %v", created, err)
	}

	// Read: every field comes back
	events, err := cal.Events(ctx, tomorrow(0), tomorrow(23))
	if err != nil || len(events) != 1 {
		t.Fatalf("events: %+v %v", events, err)
	}
	got := events[0]
	if got.Title != "Hackathon pitch" || got.Status != calendar.Tentative || got.Location != "Room 1" || got.URL == "" || !got.Start.Equal(tomorrow(9)) {
		t.Fatalf("read back: %+v", got)
	}

	// Update (same UID): confirmed, no duplicate
	got.Status = calendar.Confirmed
	got.Title = "Hackathon pitch (confirmed)"
	if _, err := cal.Add(ctx, got); err != nil {
		t.Fatal(err)
	}
	events, _ = cal.Events(ctx, tomorrow(0), tomorrow(23))
	if len(events) != 1 || events[0].Status != calendar.Confirmed {
		t.Fatalf("update: %+v", events)
	}

	// Busy
	if busy, conflicts, _ := cal.Busy(ctx, tomorrow(9).Add(15*time.Minute), tomorrow(10)); !busy || len(conflicts) != 1 {
		t.Fatal("9:15-10:00 should be busy")
	}
	if busy, _, _ := cal.Busy(ctx, tomorrow(14), tomorrow(15)); busy {
		t.Fatal("14:00-15:00 should be free")
	}

	// All-day event from JSON
	var allDay calendar.Event
	json.Unmarshal([]byte(`{"title":"Hackathon","start":"`+tomorrow(0).Format("2006-01-02")+`"}`), &allDay)
	if r, err := cal.Add(ctx, allDay); err != nil || !r.AllDay || r.End.Sub(r.Start) != 24*time.Hour {
		t.Fatalf("all-day event: %+v %v", r, err)
	}

	// Delete
	if err := cal.Delete(ctx, got.UID, ""); err != nil {
		t.Fatal(err)
	}
	events, _ = cal.Events(ctx, tomorrow(8), tomorrow(12))
	for _, e := range events {
		if e.UID == got.UID {
			t.Fatalf("still present after deletion: %+v", e)
		}
	}

	// Clear errors
	if _, err := cal.Add(ctx, calendar.Event{Title: "No date"}); err == nil {
		t.Error("expected an error without a date")
	}
	if _, err := cal.Add(ctx, calendar.Event{Title: "x", Start: tomorrow(9), Calendar: "Vacation"}); err == nil {
		t.Error("expected an error for an unknown calendar")
	}
}

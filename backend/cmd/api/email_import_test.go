package main

import (
	"context"
	"testing"
	"time"

	"github.com/dariobaldi/halendar_back/internal/data"
)

func TestParseEventExtraction(t *testing.T) {
	cases := []struct {
		name        string
		response    string
		wantOK      bool
		wantRequest bool
		wantSlots   int
	}{
		{
			name:        "clean json",
			response:    `{"requests_participation": true, "title": "Sync", "location": "", "slots": [{"date": "2026-09-20", "start": "10:00", "end": "10:30"}]}`,
			wantOK:      true,
			wantRequest: true,
			wantSlots:   1,
		},
		{
			name: "wrapped in markdown fences and prose",
			response: "Sure, here's the analysis:\n```json\n" +
				`{"requests_participation": false, "title": "", "location": "", "slots": []}` +
				"\n```\nLet me know if you need anything else.",
			wantOK:      true,
			wantRequest: false,
			wantSlots:   0,
		},
		{
			name:      "garbage, no json object",
			response:  "I cannot determine this.",
			wantOK:    false,
			wantSlots: 0,
		},
		{
			name:        "open question with no explicit slots",
			response:    `{"requests_participation": true, "title": "Catch up", "location": "", "slots": []}`,
			wantOK:      true,
			wantRequest: true,
			wantSlots:   0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			extraction, ok := parseEventExtraction(tc.response)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if extraction.RequestsParticipation != tc.wantRequest {
				t.Errorf("RequestsParticipation = %v, want %v", extraction.RequestsParticipation, tc.wantRequest)
			}
			if len(extraction.Slots) != tc.wantSlots {
				t.Errorf("len(Slots) = %d, want %d", len(extraction.Slots), tc.wantSlots)
			}
		})
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "  ", "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
	if got := firstNonEmpty("title", "fallback"); got != "title" {
		t.Errorf("got %q, want %q", got, "title")
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFirstFreeSlot(t *testing.T) {
	base := time.Date(2026, 9, 21, 10, 0, 0, 0, time.UTC)
	slots := []data.EmailEventSlot{
		{StartAt: base, EndAt: base.Add(30 * time.Minute), Availability: data.SlotAvailabilityBusy, Source: data.SlotSourceSender},
		{StartAt: base.AddDate(0, 0, 1), EndAt: base.AddDate(0, 0, 1).Add(30 * time.Minute), Availability: data.SlotAvailabilityFree, Source: data.SlotSourceSender},
	}
	got := firstFreeSlot(slots)
	if got == nil || !got.StartAt.Equal(slots[1].StartAt) {
		t.Fatalf("expected the second (free) slot, got %+v", got)
	}

	// A "free" slot from a suggested-alternative source doesn't count -- that search
	// only happens once we already know none of the sender's own times were free.
	suggestedOnly := []data.EmailEventSlot{
		{StartAt: base, EndAt: base.Add(30 * time.Minute), Availability: data.SlotAvailabilityFree, Source: data.SlotSourceSuggested},
	}
	if got := firstFreeSlot(suggestedOnly); got != nil {
		t.Fatalf("expected nil for a suggested-only slot, got %+v", got)
	}
}

func TestResolveEventSlots(t *testing.T) {
	app := &app{}
	loc := time.UTC

	got := app.resolveEventSlots(context.Background(), nil, loc, []extractedSlot{
		// Explicit, valid end -- used as-is.
		{Date: "2026-09-18", Start: "09:30", End: "10:00"},
		// Omitted end (the model correctly following the "don't guess" instruction
		// when no duration was stated) -- this is the real bug report: a message
		// like this used to have its slot silently dropped entirely, making a
		// correctly extracted date/time look like nothing had been found.
		{Date: "2026-09-18", Start: "14:00"},
		// End identical to start (the older documented model quirk) -- same
		// fallback applies rather than dropping the slot.
		{Date: "2026-09-19", Start: "11:00", End: "11:00"},
		// Unparseable start -- still correctly dropped.
		{Date: "not-a-date", Start: "09:00", End: "09:30"},
	})

	if len(got) != 3 {
		t.Fatalf("len(got) = %d, want 3 (the unparseable one dropped): %+v", len(got), got)
	}

	explicit := got[0]
	if !explicit.EndAt.Equal(time.Date(2026, 9, 18, 10, 0, 0, 0, loc)) {
		t.Errorf("explicit end not respected: %+v", explicit)
	}

	omitted := got[1]
	wantOmittedEnd := time.Date(2026, 9, 18, 14, 0, 0, 0, loc).Add(defaultSlotDuration)
	if !omitted.EndAt.Equal(wantOmittedEnd) {
		t.Errorf("omitted end = %v, want default duration applied (%v)", omitted.EndAt, wantOmittedEnd)
	}

	identical := got[2]
	wantIdenticalEnd := time.Date(2026, 9, 19, 11, 0, 0, 0, loc).Add(defaultSlotDuration)
	if !identical.EndAt.Equal(wantIdenticalEnd) {
		t.Errorf("end-equals-start = %v, want default duration applied (%v)", identical.EndAt, wantIdenticalEnd)
	}
}

// fakeCalendarSource is a minimal calendarimport.Source test double: everything is
// free except the given busy ranges.
type fakeCalendarSource struct {
	busyStart, busyEnd time.Time
	loc                *time.Location
}

func (f fakeCalendarSource) Busy(ctx context.Context, start, end time.Time) (bool, error) {
	return start.Before(f.busyEnd) && end.After(f.busyStart), nil
}

func (f fakeCalendarSource) Timezone() *time.Location { return f.loc }

func (f fakeCalendarSource) AddEvent(ctx context.Context, title, location, description string, start, end time.Time) error {
	return nil
}

func TestFindAlternativeSlots(t *testing.T) {
	loc := time.UTC
	monday := time.Date(2026, 9, 21, 10, 0, 0, 0, loc) // confirmed Monday
	proposed := []data.EmailEventSlot{
		{StartAt: monday, EndAt: monday.Add(30 * time.Minute), Availability: data.SlotAvailabilityBusy, Source: data.SlotSourceSender},
	}
	// Tuesday (the very next day) is busy too, so the first alternative should skip
	// straight to Wednesday, then Thursday, Friday -- stopping at maxAlternatives (3)
	// without ever considering the following Saturday/Sunday.
	tuesday := monday.AddDate(0, 0, 1)
	source := fakeCalendarSource{busyStart: tuesday, busyEnd: tuesday.Add(30 * time.Minute), loc: loc}

	app := &app{}
	alternatives := app.findAlternativeSlots(context.Background(), source, loc, proposed)

	if len(alternatives) != maxAlternatives {
		t.Fatalf("len(alternatives) = %d, want %d", len(alternatives), maxAlternatives)
	}
	wantDays := []int{2, 3, 4} // Wed, Thu, Fri offsets from Monday
	for i, want := range wantDays {
		got := alternatives[i]
		if !got.StartAt.Equal(monday.AddDate(0, 0, want)) {
			t.Errorf("alternative %d start = %v, want offset %d days from Monday", i, got.StartAt, want)
		}
		if got.Source != data.SlotSourceSuggested {
			t.Errorf("alternative %d source = %q, want %q", i, got.Source, data.SlotSourceSuggested)
		}
		if got.Availability != data.SlotAvailabilityFree {
			t.Errorf("alternative %d availability = %q, want %q", i, got.Availability, data.SlotAvailabilityFree)
		}
	}
}

func TestFindAlternativeSlotsNoSource(t *testing.T) {
	app := &app{}
	got := app.findAlternativeSlots(context.Background(), nil, time.UTC, []data.EmailEventSlot{
		{StartAt: time.Now(), EndAt: time.Now().Add(time.Hour)},
	})
	if got != nil {
		t.Fatalf("expected nil with no calendar source, got %+v", got)
	}
}

func TestFormatSlot(t *testing.T) {
	slot := data.EmailEventSlot{
		StartAt: time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, 9, 22, 14, 30, 0, 0, time.UTC),
	}
	got := formatSlot(slot, time.UTC)
	want := "Tuesday, Sep 22 14:00-14:30"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatSlotList(t *testing.T) {
	slots := []data.EmailEventSlot{
		{StartAt: time.Date(2026, 9, 22, 14, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 9, 22, 14, 30, 0, 0, time.UTC)},
		{StartAt: time.Date(2026, 9, 23, 9, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 9, 23, 9, 30, 0, 0, time.UTC)},
	}
	got := formatSlotList(slots, time.UTC)
	want := "Tuesday, Sep 22 14:00-14:30, Wednesday, Sep 23 09:00-09:30"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got := formatSlotList(nil, time.UTC); got != "" {
		t.Errorf("got %q, want empty for no slots", got)
	}
}

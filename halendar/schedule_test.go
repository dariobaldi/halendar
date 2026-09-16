package main

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"halendar/calendar"
	"halendar/mail"
	"halendar/testutil"
)

func TestParseSlots(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Paris")
	text := "Hi,\n\nHere are some options:\n2026-09-18 14:00-14:30\nnot a slot\n2026-09-19T09:00-09:15\n\nThanks"

	slots, err := parseSlots(text, loc)
	if err != nil {
		t.Fatal(err)
	}
	if len(slots) != 2 {
		t.Fatalf("expected 2 slots, got %d: %+v", len(slots), slots)
	}
	if !slots[0].start.Equal(time.Date(2026, 9, 18, 14, 0, 0, 0, loc)) || !slots[0].end.Equal(time.Date(2026, 9, 18, 14, 30, 0, 0, loc)) {
		t.Errorf("first slot: %+v", slots[0])
	}
	if !slots[1].start.Equal(time.Date(2026, 9, 19, 9, 0, 0, 0, loc)) || !slots[1].end.Equal(time.Date(2026, 9, 19, 9, 15, 0, 0, loc)) {
		t.Errorf("second slot: %+v", slots[1])
	}
}

func TestRunScheduleBooksFirstFreeSlot(t *testing.T) {
	ctx := context.Background()
	im := testutil.NewFakeIMAP(t)
	sm := testutil.NewFakeSMTP(t)
	cd := testutil.NewFakeCalDAV(t)
	mailbox := mail.New(mail.Config{
		IMAPHost: im.Addr, IMAPInsecure: true,
		SMTPHost: sm.Host, SMTPPort: sm.Port,
		User: im.User, Pass: im.Pass,
	})
	cal := calendar.New(calendar.Config{URL: cd.URL, User: cd.User, Pass: cd.Pass})
	loc := cal.Timezone()

	day := time.Now().In(loc).AddDate(0, 0, 1).Format("2006-01-02")
	cd.Book("evt-existing", "Already booked", parseTime(t, day+"T13:45", loc), parseTime(t, day+"T14:15", loc))

	body := "Proposed slots:\n" + day + " 14:00-14:30\n" + day + " 15:00-15:30\n"
	im.Deliver(t, "<slots@x>", "Claire <claire@example.com>", "Availabilities", body)
	msgs, err := mailbox.Recent(ctx, 1)
	if err != nil || len(msgs) != 1 {
		t.Fatalf("could not read back the seeded mail: %+v %v", msgs, err)
	}
	uid := msgs[0].UID

	if err := run(ctx, "schedule", []string{strconv.FormatUint(uint64(uid), 10)}, mailbox, cal); err != nil {
		t.Fatal(err)
	}

	// 14:00-14:30 was already busy, so the 15:00 slot should have been booked instead.
	events, err := cal.Events(ctx, parseTime(t, day+"T00:00", loc), parseTime(t, day+"T23:59", loc))
	if err != nil {
		t.Fatal(err)
	}
	var booked *calendar.Event
	for i, e := range events {
		if e.Title == "Availabilities" {
			booked = &events[i]
		}
	}
	if booked == nil {
		t.Fatalf("no event titled %q found among %+v", "Availabilities", events)
	}
	if !booked.Start.Equal(parseTime(t, day+"T15:00", loc)) {
		t.Fatalf("expected the 15:00 slot (14:00 was taken), got %s", booked.Start)
	}

	// A reply confirming the booking should have gone out in the same thread.
	sent := sm.Mails()
	if len(sent) != 1 || !strings.Contains(sent[0].Raw, "Subject: Re: Availabilities") || !strings.Contains(sent[0].Raw, "In-Reply-To: <slots@x>") {
		t.Errorf("expected one threaded reply, got: %+v", sent)
	}
}

func parseTime(t *testing.T, s string, loc *time.Location) time.Time {
	t.Helper()
	tm, err := calendar.ParseDate(s, loc)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}

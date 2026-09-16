package schedule_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"halendar/calendar"
	"halendar/mail"
	"halendar/schedule"
	"halendar/testutil"
)

func TestProposeSkipsBusySlot(t *testing.T) {
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

	// Propose must pick the free 15:00 slot, and must not book or send anything.
	proposal, err := schedule.Propose(ctx, mailbox, cal, msgs[0].UID)
	if err != nil {
		t.Fatal(err)
	}
	if !proposal.Slot.Start.Equal(parseTime(t, day+"T15:00", loc)) {
		t.Fatalf("expected the 15:00 slot (14:00 was taken), got %s", proposal.Slot.Start)
	}
	if proposal.Message == "" {
		t.Error("expected a non-empty confirmation message")
	}
	if events, _ := cal.Events(ctx, parseTime(t, day+"T00:00", loc), parseTime(t, day+"T23:59", loc)); len(events) != 1 {
		t.Fatalf("Propose must not book anything, found %d events", len(events))
	}
	if len(sm.Mails()) != 0 {
		t.Fatal("Propose must not send any reply")
	}

	// Confirm is the only step that actually books and replies.
	booked, err := schedule.Confirm(ctx, mailbox, cal, proposal)
	if err != nil {
		t.Fatal(err)
	}
	if !booked.Start.Equal(parseTime(t, day+"T15:00", loc)) {
		t.Fatalf("booked the wrong slot: %+v", booked)
	}
	sent := sm.Mails()
	if len(sent) != 1 || !strings.Contains(sent[0].Raw, "Subject: Re: Availabilities") || !strings.Contains(sent[0].Raw, "In-Reply-To: <slots@x>") {
		t.Errorf("expected one threaded reply, got: %+v", sent)
	}
}

func TestProposeNoFreeSlot(t *testing.T) {
	ctx := context.Background()
	im := testutil.NewFakeIMAP(t)
	cd := testutil.NewFakeCalDAV(t)
	mailbox := mail.New(mail.Config{IMAPHost: im.Addr, IMAPInsecure: true, User: im.User, Pass: im.Pass})
	cal := calendar.New(calendar.Config{URL: cd.URL, User: cd.User, Pass: cd.Pass})
	loc := cal.Timezone()

	day := time.Now().In(loc).AddDate(0, 0, 1).Format("2006-01-02")
	cd.Book("evt-existing", "Already booked", parseTime(t, day+"T13:45", loc), parseTime(t, day+"T14:15", loc))

	im.Deliver(t, "<slots@x>", "Claire <claire@example.com>", "Availabilities", "Only option:\n"+day+" 14:00-14:30\n")
	msgs, _ := mailbox.Recent(ctx, 1)

	if _, err := schedule.Propose(ctx, mailbox, cal, msgs[0].UID); err == nil {
		t.Error("expected an error when every proposed slot is busy")
	}
}

func TestParseSlotsViaPropose(t *testing.T) {
	ctx := context.Background()
	im := testutil.NewFakeIMAP(t)
	cd := testutil.NewFakeCalDAV(t)
	mailbox := mail.New(mail.Config{IMAPHost: im.Addr, IMAPInsecure: true, User: im.User, Pass: im.Pass})
	cal := calendar.New(calendar.Config{URL: cd.URL, User: cd.User, Pass: cd.Pass})
	loc := cal.Timezone()

	day := time.Now().In(loc).AddDate(0, 0, 1).Format("2006-01-02")
	im.Deliver(t, "<slots@x>", "Claire <claire@example.com>", "Availabilities",
		"Here are some options:\n"+day+" 14:00-14:30\nnot a slot\n"+day+"T09:00-09:15\n\nThanks")
	msgs, _ := mailbox.Recent(ctx, 1)

	proposal, err := schedule.Propose(ctx, mailbox, cal, msgs[0].UID)
	if err != nil {
		t.Fatal(err)
	}
	// The first well-formed line ("14:00-14:30") should win since nothing is booked yet.
	if !proposal.Slot.Start.Equal(parseTime(t, day+"T14:00", loc)) {
		t.Fatalf("expected the first parsed slot to be picked, got %s", proposal.Slot.Start)
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

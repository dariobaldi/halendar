package mail_test

import (
	"context"
	"strings"
	"testing"

	"halendar/mail"
	"halendar/testutil"
)

func newMailbox(t *testing.T) (*mail.Mailbox, *testutil.FakeIMAP, *testutil.FakeSMTP) {
	im := testutil.NewFakeIMAP(t)
	sm := testutil.NewFakeSMTP(t)
	mb := mail.New(mail.Config{
		IMAPHost: im.Addr, IMAPInsecure: true,
		SMTPHost: sm.Host, SMTPPort: sm.Port,
		User: im.User, Pass: im.Pass,
	})
	return mb, im, sm
}

func TestReadSearchMarkMove(t *testing.T) {
	ctx := context.Background()
	mb, im, _ := newMailbox(t)
	if err := mb.Test(ctx); err != nil {
		t.Fatal(err)
	}

	// NewSince: the first call is just a starting point, nothing is returned yet
	im.Deliver(t, "<old@x>", "old@example.com", "Old", "history")
	msgs, last, err := mb.NewSince(ctx, 0)
	if err != nil || len(msgs) != 0 || last != 1 {
		t.Fatalf("starting point: %d msgs, last=%d, err=%v", len(msgs), last, err)
	}
	im.Deliver(t, "<meeting@x>", "Claire Martin <claire@example.com>", "Project sync", "Free Thursday 2pm?")
	im.Deliver(t, "<news@x>", "news@example.com", "Newsletter", "Promo")
	msgs, last, err = mb.NewSince(ctx, last)
	if err != nil || len(msgs) != 2 || last != 3 {
		t.Fatalf("new messages: %d msgs, last=%d, err=%v", len(msgs), last, err)
	}
	if msgs, _, _ := mb.NewSince(ctx, last); len(msgs) != 0 {
		t.Fatalf("expected no new messages, got %d", len(msgs))
	}

	// Recent: newest first, with the Message-ID in angle brackets
	recent, err := mb.Recent(ctx, 2)
	if err != nil || len(recent) != 2 || recent[0].Subject != "Newsletter" || recent[1].ID != "<meeting@x>" {
		t.Fatalf("recent: %+v %v", recent, err)
	}
	claire := recent[1]
	if claire.From != "claire@example.com" || claire.FromName != "Claire Martin" || claire.Read {
		t.Fatalf("message parsed incorrectly: %+v", claire)
	}

	// Search + MarkRead
	res, err := mb.Search(ctx, mail.SearchQuery{Subject: "sync"})
	if err != nil || len(res) != 1 || res[0].UID != claire.UID {
		t.Fatalf("search by subject: %+v %v", res, err)
	}
	if err := mb.MarkRead(ctx, true, claire.UID); err != nil {
		t.Fatal(err)
	}
	unread, _ := mb.Search(ctx, mail.SearchQuery{Unread: true})
	if len(unread) != 2 {
		t.Fatalf("expected 2 unread, got %d", len(unread))
	}
	if msg, _ := mb.Read(ctx, claire.UID); !msg.Read {
		t.Fatal("the message should be marked read")
	}

	// Folders + Move
	folders, _ := mb.Folders(ctx)
	if !strings.Contains(strings.Join(folders, ","), "Archives") {
		t.Fatalf("folders: %v", folders)
	}
	if err := mb.Move(ctx, "Archives", claire.UID); err != nil {
		t.Fatal(err)
	}
	if im.Count("INBOX") != 2 || im.Count("Archives") != 1 {
		t.Fatalf("move: INBOX=%d Archives=%d", im.Count("INBOX"), im.Count("Archives"))
	}
}

func TestSendReplyDraft(t *testing.T) {
	ctx := context.Background()
	mb, im, sm := newMailbox(t)
	im.Deliver(t, "<meeting@x>", "Claire <claire@example.com>", "Project sync", "Free Thursday?")
	msgs, _ := mb.Recent(ctx, 1)

	// Reply: recipient, Re:, thread headers
	if _, err := mb.Reply(ctx, msgs[0], "Thursday 2pm works for me."); err != nil {
		t.Fatal(err)
	}
	// Send with a copy and HTML
	outgoing := mail.Outgoing{To: []string{"a@x.com"}, Cc: []string{"b@x.com"}, Subject: "Summer ✓", Text: "Hello", HTML: "<p>Hello</p>"}
	if _, err := mb.Send(ctx, outgoing); err != nil {
		t.Fatal(err)
	}
	sent := sm.Mails()
	if len(sent) != 2 {
		t.Fatalf("expected 2 mails, got %d", len(sent))
	}
	raw := sent[0].Raw
	for _, want := range []string{"To: claire@example.com", "Subject: Re: Project sync", "In-Reply-To: <meeting@x>", "References: <meeting@x>"} {
		if !strings.Contains(raw, want) {
			t.Errorf("reply: %q missing from\n%s", want, raw)
		}
	}
	if len(sent[1].To) != 2 || !strings.Contains(sent[1].Raw, "multipart/alternative") || !strings.Contains(sent[1].Raw, "=?utf-8?q?") {
		t.Errorf("HTML/copy send incorrect: %+v", sent[1])
	}

	// Clear error when there is no recipient
	if _, err := mb.Send(ctx, mail.Outgoing{Text: "x"}); err == nil {
		t.Error("expected an error without a recipient")
	}

	// Draft: saved to the \Drafts folder, nothing is sent
	folder, err := mb.SaveDraft(ctx, mail.ReplyTo(msgs[0], "To review"))
	if err != nil || folder != "Drafts" || im.Count("Drafts") != 1 || len(sm.Mails()) != 2 {
		t.Fatalf("draft: folder=%q err=%v drafts=%d sent=%d", folder, err, im.Count("Drafts"), len(sm.Mails()))
	}
}

func TestReplyRoutesToReplyToWhenPresent(t *testing.T) {
	ctx := context.Background()
	mb, im, sm := newMailbox(t)

	// NewSince(ctx, 0) always just establishes a starting point, never returning
	// anything (see TestReadSearchMarkMove) -- deliver one message first so the
	// baseline UID it captures is non-zero, or the very next call would be
	// mistaken for another 0-baseline call instead of a real "what's new" check.
	im.Deliver(t, "<old@x>", "old@example.com", "Old", "history")
	_, last, err := mb.NewSince(ctx, 0)
	if err != nil {
		t.Fatal(err)
	}

	// A Reply-To distinct from From (the common newsletter/automated-sender case) --
	// the reply must go there, but the message should still display as being from
	// the real sender, not silently swapped for the Reply-To address.
	im.DeliverWithReplyTo(t, "<promo@x>", "Acme Team <noreply@acme.com>", "support@acme.com", "Big sale", "50% off everything")
	msgs, last, err := mb.NewSince(ctx, last)
	if err != nil || len(msgs) != 1 {
		t.Fatalf("delivering: %d msgs, err=%v", len(msgs), err)
	}
	msg := msgs[0]
	if msg.From != "noreply@acme.com" || msg.FromName != "Acme Team" {
		t.Fatalf("display sender changed by Reply-To: From=%q FromName=%q", msg.From, msg.FromName)
	}
	if msg.ReplyTo != "support@acme.com" {
		t.Fatalf("ReplyTo not captured: %q", msg.ReplyTo)
	}

	if _, err := mb.Reply(ctx, msg, "Thanks!"); err != nil {
		t.Fatal(err)
	}
	sent := sm.Mails()
	if len(sent) != 1 || len(sent[0].To) != 1 || sent[0].To[0] != "support@acme.com" {
		t.Fatalf("reply should go to Reply-To, got %+v", sent)
	}

	// No Reply-To at all -- falls back to From, as before.
	im.Deliver(t, "<plain@x>", "plain@example.com", "Hi", "just saying hi")
	msgs, _, err = mb.NewSince(ctx, last)
	if err != nil || len(msgs) != 1 {
		t.Fatalf("delivering: %d msgs, err=%v", len(msgs), err)
	}
	if _, err := mb.Reply(ctx, msgs[0], "Hi back!"); err != nil {
		t.Fatal(err)
	}
	sent = sm.Mails()
	if len(sent) != 2 || len(sent[1].To) != 1 || sent[1].To[0] != "plain@example.com" {
		t.Fatalf("reply without Reply-To should fall back to From, got %+v", sent)
	}
}

func TestConfigDerivesSMTP(t *testing.T) {
	cfg := mail.New(mail.Config{IMAPHost: "imap.gmail.com:993", User: "me@gmail.com", Pass: "x"}).Config()
	if cfg.SMTPHost != "smtp.gmail.com" || cfg.SMTPPort != 587 || cfg.From != "me@gmail.com" || cfg.IMAPInsecure {
		t.Fatalf("config: %+v", cfg)
	}
}

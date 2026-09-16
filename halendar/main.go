// Command-line demo of the mail and calendar packages.
// Credentials go in the .env file (see .env.example).
//
//	go run . health                  tests the mail and calendar connection
//	go run . mails [n]                n most recent mails
//	go run . unread                   unread mails
//	go run . read <uid>               one full mail
//	go run . send send.json           sends a mail
//	go run . reply <uid> "text"       replies in the same thread
//	go run . draft send.json          saves the mail to Drafts
//	go run . calendar [days]          upcoming schedule
//	go run . busy <start> <end>       checks whether a slot is free
//	go run . add event.json           adds (or updates) one or more events
//	go run . delete <uid>             deletes an event
//	go run . schedule <uid>           proposes a slot from a mail; books and replies on confirmation
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"halendar/calendar"
	"halendar/envfile"
	"halendar/mail"
)

func main() {
	if err := envfile.Load(".env"); err != nil {
		fmt.Fprintln(os.Stderr, "✗", err)
		os.Exit(1)
	}
	mailbox := mail.New(mail.ConfigFromEnv())
	cal := calendar.New(calendar.ConfigFromEnv())

	if len(os.Args) < 2 {
		fmt.Println(helpText)
		return
	}
	if err := run(context.Background(), os.Args[1], os.Args[2:], mailbox, cal); err != nil {
		fmt.Fprintln(os.Stderr, "✗", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd string, args []string, mailbox *mail.Mailbox, cal *calendar.Client) error {
	switch cmd {
	case "health":
		return runHealth(ctx, mailbox, cal)
	case "mails":
		return runMails(ctx, mailbox, args)
	case "unread":
		return runUnread(ctx, mailbox)
	case "read":
		return runRead(ctx, mailbox, args)
	case "send", "draft":
		return runSendOrDraft(ctx, mailbox, cmd, args)
	case "reply":
		return runReply(ctx, mailbox, args)
	case "calendar":
		return runCalendar(ctx, cal, args)
	case "busy":
		return runBusy(ctx, cal, args)
	case "add":
		return runAdd(ctx, cal, args)
	case "delete":
		return runDelete(ctx, cal, args)
	case "schedule":
		return runSchedule(ctx, mailbox, cal, args)
	}
	return fmt.Errorf("unknown command %q\n\n%s", cmd, helpText)
}

func runHealth(ctx context.Context, mailbox *mail.Mailbox, cal *calendar.Client) error {
	ok := true
	if err := mailbox.Test(ctx); err != nil {
		fmt.Println("✗ mail    ", err)
		ok = false
	} else {
		n, _ := mailbox.Count(ctx)
		fmt.Printf("✓ mail     %s (%d messages) · sends via %s:%d\n", mailbox.Config().User, n, mailbox.Config().SMTPHost, mailbox.Config().SMTPPort)
	}
	if err := cal.Test(ctx); err != nil {
		fmt.Println("✗ calendar", err)
		ok = false
	} else {
		names, _ := cal.Calendars(ctx)
		fmt.Println("✓ calendar", strings.Join(names, ", "))
	}
	if !ok {
		return errors.New("at least one module is not responding")
	}
	return nil
}

func runMails(ctx context.Context, mailbox *mail.Mailbox, args []string) error {
	msgs, err := mailbox.Recent(ctx, intArg(args, 5))
	if err != nil {
		return err
	}
	printMessages(msgs)
	return nil
}

func runUnread(ctx context.Context, mailbox *mail.Mailbox) error {
	msgs, err := mailbox.Search(ctx, mail.SearchQuery{Unread: true, Max: 20})
	if err != nil {
		return err
	}
	printMessages(msgs)
	return nil
}

func runRead(ctx context.Context, mailbox *mail.Mailbox, args []string) error {
	uid, err := uidArg(args)
	if err != nil {
		return err
	}
	msg, err := mailbox.Read(ctx, uid)
	if err != nil {
		return err
	}
	return printJSON(msg)
}

func runSendOrDraft(ctx context.Context, mailbox *mail.Mailbox, cmd string, args []string) error {
	var outgoing mail.Outgoing
	if err := readJSON(args, &outgoing); err != nil {
		return err
	}
	if cmd == "draft" {
		folder, err := mailbox.SaveDraft(ctx, outgoing)
		if err != nil {
			return err
		}
		fmt.Printf("✓ draft saved to %q\n", folder)
		return nil
	}
	id, err := mailbox.Send(ctx, outgoing)
	if err != nil {
		return err
	}
	fmt.Printf("✓ mail sent to %s (%s)\n", strings.Join(outgoing.To, ", "), id)
	return nil
}

func runReply(ctx context.Context, mailbox *mail.Mailbox, args []string) error {
	uid, err := uidArg(args)
	if err != nil || len(args) < 2 {
		return errors.New(`usage: go run . reply <uid> "text"`)
	}
	msg, err := mailbox.Read(ctx, uid)
	if err != nil {
		return err
	}
	if _, err := mailbox.Reply(ctx, *msg, args[1]); err != nil {
		return err
	}
	fmt.Printf("✓ reply sent to %s (\"Re: %s\")\n", msg.From, msg.Subject)
	return nil
}

func runCalendar(ctx context.Context, cal *calendar.Client, args []string) error {
	start := time.Now()
	end := start.AddDate(0, 0, intArg(args, 7))
	events, err := cal.Events(ctx, start, end)
	if err != nil {
		return err
	}
	fmt.Printf("%d event(s) from %s to %s\n", len(events), start.Format("Jan 2"), end.Format("Jan 2"))
	day := ""
	for _, e := range events {
		if d := e.Start.Format("Mon Jan 2"); d != day {
			day = d
			fmt.Println("\n" + day)
		}
		hours := e.Start.Format("15:04") + "-" + e.End.Format("15:04")
		if e.AllDay {
			hours = "all day   "
		}
		fmt.Printf("  %s  %s  [%s · %s]  uid=%s\n", hours, e.Title, e.Calendar, e.Status, e.UID)
	}
	return nil
}

func runAdd(ctx context.Context, cal *calendar.Client, args []string) error {
	raw, err := readInput(args)
	if err != nil {
		return err
	}
	var events []calendar.Event
	if trimmed := bytes.TrimSpace(raw); len(trimmed) > 0 && trimmed[0] == '[' {
		err = json.Unmarshal(trimmed, &events)
	} else {
		var e calendar.Event
		err = json.Unmarshal(trimmed, &e)
		events = []calendar.Event{e}
	}
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	succeeded := 0
	for i, e := range events {
		added, err := cal.Add(ctx, e)
		if err != nil {
			fmt.Printf("✗ event %d (%q): %v\n", i+1, e.Title, err)
			continue
		}
		succeeded++
		fmt.Printf("✓ %q → %s, on %s from %s to %s (%s)\n", added.Title, added.Calendar,
			added.Start.Format("Jan 2, 2006"), added.Start.Format("15:04"), added.End.Format("15:04"), added.Status)
	}
	if succeeded < len(events) {
		return fmt.Errorf("%d/%d event(s) added", succeeded, len(events))
	}
	return nil
}

func runDelete(ctx context.Context, cal *calendar.Client, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: go run . delete <uid> [calendar]")
	}
	calendarName := ""
	if len(args) > 1 {
		calendarName = args[1]
	}
	if err := cal.Delete(ctx, args[0], calendarName); err != nil {
		return err
	}
	fmt.Println("✓ deleted")
	return nil
}

func printMessages(msgs []mail.Message) {
	if len(msgs) == 0 {
		fmt.Println("no mails")
	}
	for _, msg := range msgs {
		unread := "●"
		if msg.Read {
			unread = " "
		}
		fmt.Println(strings.Repeat("─", 60))
		fmt.Printf("%s UID %d · %s\n  From    %s %s\n  Subject %s\n\n%s\n",
			unread, msg.UID, msg.Date.Format("Jan 2 15:04"), msg.FromName, msg.From, msg.Subject, truncate(msg.Text, 400))
	}
}

func readInput(args []string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.New("missing JSON file (or \"-\" for standard input)")
	}
	if args[0] == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(args[0])
}

func readJSON(args []string, v any) error {
	raw, err := readInput(args)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func uidArg(args []string) (uint32, error) {
	if len(args) == 0 {
		return 0, errors.New("missing UID")
	}
	n, err := strconv.ParseUint(args[0], 10, 32)
	return uint32(n), err
}

func intArg(args []string, fallback int) int {
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

const helpText = `Commands:
  go run . health                  tests the mail and calendar connection
  go run . mails [n]                n most recent mails
  go run . unread                   unread mails
  go run . read <uid>               one full mail (JSON)
  go run . send send.json           sends a mail
  go run . reply <uid> "text"       replies in the same thread
  go run . draft send.json          saves the mail to Drafts
  go run . calendar [days]          upcoming schedule
  go run . busy <start> <end>       checks whether a slot is free
  go run . add event.json           adds (or updates) one or more events
  go run . delete <uid>             deletes an event
  go run . schedule <uid>           proposes a slot from a mail; books and replies on confirmation`

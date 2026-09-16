package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"halendar/calendar"
	"halendar/mail"
)

// slotLine matches one proposed time slot per line, e.g. "2026-09-18 14:00-14:30".
var slotLine = regexp.MustCompile(`(\d{4}-\d{2}-\d{2})[ T](\d{2}:\d{2})-(\d{2}:\d{2})`)

type slot struct {
	start, end time.Time
}

// runSchedule reads a mail proposing time slots (one per line, see slotLine), books the
// first slot that is free on the calendar, and replies to the mail confirming it.
func runSchedule(ctx context.Context, mailbox *mail.Mailbox, cal *calendar.Client, args []string) error {
	uid, err := uidArg(args)
	if err != nil {
		return err
	}
	msg, err := mailbox.Read(ctx, uid)
	if err != nil {
		return err
	}

	slots, err := parseSlots(msg.Text, cal.Timezone())
	if err != nil {
		return err
	}
	if len(slots) == 0 {
		return errors.New("no time slot found in the mail body (expected lines like 2026-09-18 14:00-14:30)")
	}

	for _, s := range slots {
		busy, _, err := cal.Busy(ctx, s.start, s.end)
		if err != nil {
			return err
		}
		if busy {
			continue
		}

		booked, err := cal.Add(ctx, calendar.Event{
			Title:       msg.Subject,
			Start:       s.start,
			End:         s.end,
			Description: "Booked from a mail sent by " + msg.From,
		})
		if err != nil {
			return err
		}
		confirmation := confirmationText(booked)
		if _, err := mailbox.Reply(ctx, *msg, confirmation); err != nil {
			return err
		}
		fmt.Printf("✓ booked %s–%s on %s (uid=%s)\n", booked.Start.Format("15:04"), booked.End.Format("15:04"), booked.Start.Format("Mon Jan 2"), booked.UID)
		fmt.Printf("✓ replied to %s: %q\n", msg.From, confirmation)
		return nil
	}
	return fmt.Errorf("none of the %d proposed slot(s) are free", len(slots))
}

// parseSlots extracts every slotLine match from text and turns it into a slot in loc.
func parseSlots(text string, loc *time.Location) ([]slot, error) {
	var slots []slot
	for _, line := range strings.Split(text, "\n") {
		match := slotLine.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		date, startTime, endTime := match[1], match[2], match[3]
		start, err := calendar.ParseDate(date+"T"+startTime, loc)
		if err != nil {
			return nil, fmt.Errorf("slot %q: %w", strings.TrimSpace(line), err)
		}
		end, err := calendar.ParseDate(date+"T"+endTime, loc)
		if err != nil {
			return nil, fmt.Errorf("slot %q: %w", strings.TrimSpace(line), err)
		}
		slots = append(slots, slot{start: start, end: end})
	}
	return slots, nil
}

// confirmationText is a plain, generic confirmation message. Once the backend's LLM step
// exists, its generated message can be passed to mailbox.Reply instead — the booking logic
// above does not need to change.
func confirmationText(e calendar.Event) string {
	return fmt.Sprintf("Booked: %s · %s-%s", e.Start.Format("Monday, Jan 2"), e.Start.Format("15:04"), e.End.Format("15:04"))
}

func runBusy(ctx context.Context, cal *calendar.Client, args []string) error {
	if len(args) < 2 {
		return errors.New(`usage: go run . busy <start> <end>  (e.g. "2026-09-18T14:00" "2026-09-18T14:30")`)
	}
	start, err := calendar.ParseDate(args[0], cal.Timezone())
	if err != nil {
		return err
	}
	end, err := calendar.ParseDate(args[1], cal.Timezone())
	if err != nil {
		return err
	}

	busy, conflicts, err := cal.Busy(ctx, start, end)
	if err != nil {
		return err
	}
	if !busy {
		fmt.Println("✓ free")
		return nil
	}
	fmt.Printf("✗ busy (%d conflict(s))\n", len(conflicts))
	for _, c := range conflicts {
		fmt.Printf("  %s-%s  %s [%s]\n", c.Start.Format("15:04"), c.End.Format("15:04"), c.Title, c.Calendar)
	}
	return nil
}

// Package schedule turns a mail proposing time slots into a calendar booking,
// as two separate steps so a caller (CLI, backend, ...) can show the user what
// would happen before anything is actually booked or sent.
//
//	proposal, _ := schedule.Propose(ctx, mailbox, cal, uid)  // read-only
//	// ... show proposal.Event and proposal.Message to the user ...
//	booked, _ := schedule.Confirm(ctx, mailbox, cal, proposal)  // only on "send"
package schedule

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

// Slot is one candidate time range parsed out of a mail.
type Slot struct {
	Start, End time.Time
}

// Proposal is what Propose found: the mail it read, the first free slot, the event
// that would be created, and the message that would be sent to confirm it. Nothing
// in a Proposal has been booked or sent yet.
type Proposal struct {
	Mail    mail.Message
	Slot    Slot
	Event   calendar.Event
	Message string
}

// Propose reads the mail at uid, parses candidate slots from its body (lines like
// "2026-09-18 14:00-14:30"), and returns the first one that is free on the calendar.
// It does not book anything or send any reply.
func Propose(ctx context.Context, mailbox *mail.Mailbox, cal *calendar.Client, uid uint32) (*Proposal, error) {
	msg, err := mailbox.Read(ctx, uid)
	if err != nil {
		return nil, err
	}

	slots, err := parseSlots(msg.Text, cal.Timezone())
	if err != nil {
		return nil, err
	}
	if len(slots) == 0 {
		return nil, errors.New("no time slot found in the mail body (expected lines like 2026-09-18 14:00-14:30)")
	}

	for _, s := range slots {
		busy, _, err := cal.Busy(ctx, s.Start, s.End)
		if err != nil {
			return nil, err
		}
		if busy {
			continue
		}
		event := calendar.Event{
			Title:       msg.Subject,
			Start:       s.Start,
			End:         s.End,
			Description: "Booked from a mail sent by " + msg.From,
		}
		return &Proposal{Mail: *msg, Slot: s, Event: event, Message: confirmationText(event)}, nil
	}
	return nil, fmt.Errorf("none of the %d proposed slot(s) are free", len(slots))
}

// Confirm books the proposed event and replies to the original mail with the
// proposal's message. Call it only once the user has approved the proposal.
func Confirm(ctx context.Context, mailbox *mail.Mailbox, cal *calendar.Client, p *Proposal) (calendar.Event, error) {
	booked, err := cal.Add(ctx, p.Event)
	if err != nil {
		return booked, err
	}
	if _, err := mailbox.Reply(ctx, p.Mail, p.Message); err != nil {
		return booked, err
	}
	return booked, nil
}

// parseSlots extracts every slotLine match from text and turns it into a Slot in loc.
func parseSlots(text string, loc *time.Location) ([]Slot, error) {
	var slots []Slot
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
		slots = append(slots, Slot{Start: start, End: end})
	}
	return slots, nil
}

// confirmationText is a plain, generic confirmation message. Once a backend LLM step
// exists, its generated message can be used to build the Proposal.Message instead —
// Propose and Confirm above do not need to change.
func confirmationText(e calendar.Event) string {
	return fmt.Sprintf("Booked: %s · %s-%s", e.Start.Format("Monday, Jan 2"), e.Start.Format("15:04"), e.End.Format("15:04"))
}

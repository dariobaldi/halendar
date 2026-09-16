package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"halendar/calendar"
	"halendar/mail"
	"halendar/schedule"
)

// runSchedule shows the proposed booking and only books it once the user confirms —
// the CLI equivalent of "show the prefilled details, send only on button press".
// The same Propose/Confirm split is what a backend would call from two separate
// HTTP endpoints instead of a terminal prompt.
func runSchedule(ctx context.Context, mailbox *mail.Mailbox, cal *calendar.Client, args []string) error {
	uid, err := uidArg(args)
	if err != nil {
		return err
	}
	proposal, err := schedule.Propose(ctx, mailbox, cal, uid)
	if err != nil {
		return err
	}

	fmt.Printf("Proposed: %s from %s to %s, with %s\n", proposal.Event.Title,
		proposal.Slot.Start.Format("Mon Jan 2 15:04"), proposal.Slot.End.Format("15:04"), proposal.Mail.From)
	fmt.Printf("Reply that would be sent: %q\n", proposal.Message)

	if !confirmed() {
		fmt.Println("cancelled, nothing booked")
		return nil
	}

	booked, err := schedule.Confirm(ctx, mailbox, cal, proposal)
	if err != nil {
		return err
	}
	fmt.Printf("✓ booked %s–%s on %s (uid=%s)\n", booked.Start.Format("15:04"), booked.End.Format("15:04"), booked.Start.Format("Mon Jan 2"), booked.UID)
	fmt.Printf("✓ replied to %s\n", proposal.Mail.From)
	return nil
}

func confirmed() bool {
	fmt.Print("Send? [y/N] ")
	answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.EqualFold(strings.TrimSpace(answer), "y")
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

# Halendar

A small Go toolkit for reading/sending mail over IMAP/SMTP and managing a
CalDAV calendar, plus a CLI that drives both. Three independent packages —
import them into another program, or use `main.go` as-is.

## Install

```bash
cp .env.example .env    # fill in your mail + calendar credentials
go mod tidy
go run . health          # confirms both connections work
```

## CLI

| Command | Does |
|---|---|
| `health` | tests the mail + calendar connection |
| `mails [n]` | n most recent mails |
| `unread` | unread mails |
| `read <uid>` | one full mail, as JSON |
| `send send.json` | sends a mail |
| `reply <uid> "text"` | replies in the same thread |
| `draft send.json` | saves to Drafts instead of sending |
| `calendar [days]` | upcoming schedule |
| `busy <start> <end>` | checks whether a slot is free |
| `add event.json` | adds or updates one or more events |
| `delete <uid> [calendar]` | deletes an event |
| `schedule <uid>` | books the first free slot proposed in a mail, replies confirming it |

Sample payloads live in [`examples/`](examples/).

`schedule` looks for lines like `2026-09-18 14:00-14:30` in the mail body, checks
each against the calendar with `Busy`, books the first free one, and replies in
the same thread with a plain confirmation message — no AI involved yet. That
message is the one seam meant to be swapped out later for a backend-generated
one, without touching the booking logic.

## Packages

- **`mail`** — `Mailbox`: read (`Recent`, `Search`, `Read`, `NewSince`), send (`Send`, `Reply`, `SaveDraft`), and manage (`MarkRead`, `Move`, `Folders`) a mail account.
- **`calendar`** — `Client`: `Events`, `Busy`, `Add`, `Delete` against a CalDAV calendar. `Event` decodes directly from JSON (`duration_minutes`, all-day dates, timezone).
- **`envfile`** — loads `.env` into the process environment; nothing fancier.
- **`testutil`** — fake IMAP/SMTP/CalDAV servers used by the tests below.

## Test

```bash
go test ./...
```

No real account needed — everything runs against `testutil`'s fake servers.

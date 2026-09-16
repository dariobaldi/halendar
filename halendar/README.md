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
| `add event.json` | adds or updates one or more events |
| `delete <uid> [calendar]` | deletes an event |

Sample payloads live in [`examples/`](examples/).

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

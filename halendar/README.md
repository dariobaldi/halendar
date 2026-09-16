# Halendar — mail and calendar modules

Bundles what used to be tested separately (IMAP reading, SMTP sending, CalDAV calendar)
into **two reusable Go packages** that work the same way:

```go
envfile.Load(".env")

mailbox := mail.New(mail.ConfigFromEnv())       // or mail.Config{...}
cal     := calendar.New(calendar.ConfigFromEnv()) // or calendar.Config{...}

mailbox.Test(ctx)
cal.Test(ctx)
```

## Getting started

```bash
cp .env.example .env    # credentials for the mail account and the calendar
go mod tidy
go run . health
```

## Package `mail`

| Function | Role |
|---|---|
| `Recent(ctx, n)` | n most recent mails (newest first) |
| `NewSince(ctx, lastUID)` | mails received since the last call (returns the new UID to keep) |
| `Read(ctx, uid)` | one mail: sender, recipients, text, HTML, attachments, read/unread |
| `Search(ctx, SearchQuery{...})` | unread, sender, subject, contents, dates |
| `MarkRead(ctx, read, uids...)` | read / unread |
| `Move(ctx, folder, uids...)` | move a mail |
| `Folders(ctx)` · `Count(ctx)` | the account's folders · number of mails |
| `Send(ctx, Outgoing{...})` | to / cc / bcc, text and optional HTML |
| `Reply(ctx, message, text)` | reply in the same thread (Re:, In-Reply-To, References) |
| `ReplyTo(message, text)` | prepares a reply without sending it |
| `SaveDraft(ctx, Outgoing{...})` | saves to Drafts instead of sending |

## Package `calendar`

| Function | Role |
|---|---|
| `Calendars(ctx)` | names of the calendars (task-only calendars like "Reminders" are skipped) |
| `Events(ctx, start, end)` | every event in the period, recurring events expanded |
| `Busy(ctx, start, end)` | is the slot taken? and by what |
| `Add(ctx, Event{...})` | creates, or updates if the UID already exists (status confirmed / tentative / cancelled) |
| `Delete(ctx, uid, calendar)` | deletes an event |

`calendar.Event` can be read directly from JSON:

```json
{ "title": "Pitch", "start": "2026-09-18T09:00", "duration_minutes": 60, "status": "tentative" }
```

(`"start": "2026-09-19"` means an all-day event; `"end"` can replace `"duration_minutes"`; `"timezone"` is optional.)

## Command-line demo

```bash
go run . health
go run . mails 5
go run . unread
go run . read 42
go run . send examples/send.json
go run . reply 42 "Thursday 2pm works for me."
go run . draft examples/send.json
go run . calendar 7
go run . add examples/events.json
go run . delete evt-1234abcd
```

## Tests

```bash
go test ./...
```

The tests use fake IMAP, SMTP, and CalDAV servers (`testutil/`): no real account needed.

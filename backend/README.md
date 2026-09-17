# Halendar backend

A Go REST + WebSocket API: imports mail, classifies meeting requests with an
LLM, checks calendar availability, and drafts/sends replies.

See the [root README](../README.md) for the overall architecture and how this
fits with the frontend. This doc is the reference for the backend itself.

## Stack

- **Go**, [`httprouter`](https://github.com/julienschmidt/httprouter) for
  routing.
- **Postgres**, migrated with [`golang-migrate`](https://github.com/golang-migrate/migrate).
- **AI classification**: a local [Ollama](https://ollama.com) model by
  default (`internal/ollama`), or per-user Claude/Gemini API keys
  (`internal/claude`, `internal/gemini`) as an opt-in alternative.
- **Mail**: `halendar/mail` (a sibling Go module, see the root README) for
  IMAP/SMTP; Gmail additionally supports OAuth-based import
  (`internal/emailimport`).
- **Calendar**: Google Calendar OAuth or CalDAV (`internal/calendarimport`,
  `halendar/calendar`).
- **Push notifications**: Firebase Cloud Messaging (`internal/push`).
- **Realtime**: a WebSocket endpoint (`internal/websocket`) pushes
  `{"type": ..., "refresh": true}` events so the frontend can invalidate
  and refetch instead of polling.
- Secrets (OAuth refresh tokens, app passwords, AI API keys) are encrypted at
  rest with AES-256-GCM (`internal/secretbox`), keyed by `ENCRYPTION_KEY`.

## Directory structure

```
cmd/api/          The API server. One file per resource group (proposals.go,
                   email_accounts.go, calendar.go, ...); routes.go wires them up.
cmd/bootstrap/     CLI to create the first user (see "First user" below).
internal/data/     Postgres models -- one file per table/resource, plus models.go's
                   Models{} aggregate.
internal/          Everything else: one package per integration (see Stack above).
migrations/        golang-migrate SQL migrations, applied in order.
remote/            Production deploy: setup/01.sh (provisions a fresh VPS),
                   production/ (Caddyfile + systemd unit templates), backup/.
docs/              postman_collection.json -- every endpoint, ready to import.
```

## Running locally

From the repo root, `make setup` creates `backend/.env` interactively and
brings up Postgres/Ollama/Adminer + migrations. Then, from `backend/`:

```bash
go run ./cmd/api          # starts on :4355 (PORT + HOST_IP in .env)
```

For local dev the API runs natively (not inside `docker-compose`) so it can
be rebuilt/restarted instantly; it reaches Postgres and Ollama via the
host-mapped ports `docker-compose.yml` publishes for exactly this reason
(`DB_DSN_DEV`, `OLLAMA_BASE_URL` -- both set correctly by `make setup`).

### First user

`POST /v1/users` (normal registration) requires an existing admin, so a fresh
database has no way to create a user through the API. Run once after
migrations:

```bash
go run ./cmd/bootstrap    # prompts for name/email/username/password
# or: make bootstrap-admin
```

This creates a fully-activated user at the top access level (`DevLevel`, see
`cmd/api/routes.go`).

## Environment variables

Everything below lives in `backend/.env` (gitignored; `make setup` /
`scripts/setup-env.sh` generates it, and `backend/.envCONFIG` is the
annotated template it's generated from). Anything marked optional degrades
the corresponding feature cleanly instead of failing to start.

| Variable | Required? | What it's for |
|---|---|---|
| `PORT`, `HOST_IP`, `FRONTEND_URL`, `TRUSTED_ORIGINS`, `ENV` | Required | Server binding, CORS, and where OAuth flows redirect back to |
| `DATABASE_NAME`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `DB_PORT`, `DB_DSN`, `DB_DSN_DEV` | Required | Postgres connection (`DB_DSN_DEV` is used whenever `ENV != production`) |
| `ENCRYPTION_KEY` | Required to connect email/calendar accounts | AES-256-GCM key for encrypted secrets at rest (`openssl rand -base64 32`) |
| `SMTP_*` | Optional | Transactional email (activation/welcome mail) |
| `MAIL_*` | Optional | The mailbox `/v1/mail/*` reads/sends through -- a Gmail app password or any IMAP/SMTP account |
| `CALDAV_*` | Optional | A CalDAV calendar (e.g. iCloud) for availability checks |
| `OLLAMA_MODEL`, `OLLAMA_BASE_URL` | Optional (has defaults) | Local classification model. `gemma3:4b` is the tuned default; `gemma3:1b` is lighter but less reliable (see the doc comment on `eventExtractionPromptTmpl`) |
| `CLAUDE_MODEL`, `GEMINI_MODEL` | Optional (has defaults) | Model used *if* a user connects their own Claude/Gemini key in Settings -- no key lives in `.env` |
| `FCM_PROJECT_ID`, `FCM_SERVICE_ACCOUNT_FILE` | Optional | Push notifications -- service account JSON from Firebase Console → Project Settings → Service Accounts |
| `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, `GOOGLE_OAUTH_REDIRECT_URL`, `GOOGLE_OAUTH_CALENDAR_REDIRECT_URL` | Optional | Gmail import + Google Calendar connect (one OAuth client, both redirect URLs registered on it) |
| `EMAIL_SYNC_INTERVAL` | Optional (default `1m`) | How often connected mail accounts are polled |

## Commands

`make help` lists everything with a one-line description. Highlights:

```
make bootstrap-admin        create the first user
make migrate/create name=X  new migration
make migrate/up              apply pending migrations
make audit                   go fmt + govulncheck + staticcheck + go test -race
make build/api                cross-compile the Linux binary used in deploys
make connect                  ssh to the production server
make deploy                   apply migrations, sync Caddy, rebuild+redeploy the API container
make db/backup / db/import    round-trip a production DB dump to/from local Postgres
```

## Testing

```bash
go test -race ./...
```

`cmd/api/email_import_test.go` covers the AI-response parsing and slot
resolution/formatting logic (`TestParseEventExtraction`,
`TestResolveEventSlots`, `TestFindAlternativeSlots`, ...) without needing a
database. `halendar/mail/mail_test.go` (sibling module) covers IMAP/SMTP
behavior. `go vet ./...` and `staticcheck ./...` (via `make audit`) are the
static checks currently run; there's no CI wired up to run any of this
automatically yet.

## Deploying to your own server

Optional, and separate from local dev setup. From the repo root:

```bash
make setup-deploy              # interactive: domain, TLS email, VPS details
make -C backend deploy/setup   # provisions the server (backend/remote/setup/01.sh) -- ends in a reboot
make -C backend deploy         # migrations + Caddy + API container
```

`remote/setup/01.sh` installs Docker, Caddy, the `migrate` CLI, and `fail2ban`
on a fresh Ubuntu box, creates a non-root SSH-key-only deploy user, and
creates the Docker volumes `docker-compose.yml` expects to already exist
(`halendar-db`, `halendar-api`, `halendar-ollama`, `config_files` --
`docker-compose.yml` marks them `external: true`, i.e. it expects them
pre-created, same as local dev's `scripts/setup-dev.sh` does).

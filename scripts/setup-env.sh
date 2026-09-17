#!/usr/bin/env bash
# Interactively creates the local config files a fresh clone of Halendar
# needs to run: backend/.env, frontend/lib/env.dart, and (optionally)
# the Firebase credential files used for push notifications.
#
# Safe to re-run: any file that already exists is left untouched. Run this
# on its own with `bash scripts/setup-env.sh`, or via `make setup`.
set -euo pipefail
cd "$(dirname "$0")/.."

BACKEND_ENV="backend/.env"
BACKEND_ENV_TEMPLATE="backend/.envCONFIG"
FRONTEND_ENV="frontend/lib/env.dart"
FIREBASE_SA="backend/firebase-service-account.json"
GOOGLE_SERVICES="frontend/android/app/google-services.json"

bold() { printf '\033[1m%s\033[0m\n' "$*"; }
section() { echo; bold "== $* =="; }
info() { printf '  %s\n' "$*"; }
skip_note() { echo "  ${1} already exists -- leaving it untouched."; }

# ask VAR "question" "default"
ask() {
  local __varname="$1" __question="$2" __default="${3-}" __answer
  if [ -n "$__default" ]; then
    read -r -p "  ${__question} [${__default}]: " __answer
  else
    read -r -p "  ${__question}: " __answer
  fi
  printf -v "$__varname" '%s' "${__answer:-$__default}"
}

ask_hidden() {
  local __varname="$1" __question="$2" __answer
  read -r -s -p "  ${__question}: " __answer
  echo
  printf -v "$__varname" '%s' "$__answer"
}

# Rewrites $BACKEND_ENV_TEMPLATE into $BACKEND_ENV, replacing only the KEY=
# lines named in $BACKEND_ENV_SUBST_KEYS with the matching value from the
# environment -- everything else (including the doc comments already written
# in .envCONFIG, and the ${VAR} interpolation godotenv resolves at load time
# for DB_DSN/DB_DSN_DEV) passes through unchanged.
render_backend_env() {
  awk -v keys="$BACKEND_ENV_SUBST_KEYS" '
    BEGIN { n = split(keys, arr, " "); for (i = 1; i <= n; i++) allow[arr[i]] = 1 }
    {
      if (match($0, /^[A-Z_][A-Z0-9_]*=/)) {
        key = substr($0, 1, RLENGTH - 1)
        if (key in allow && key in ENVIRON) {
          print key "=\"" ENVIRON[key] "\""
          next
        }
      }
      print $0
    }
  ' "$BACKEND_ENV_TEMPLATE" > "$BACKEND_ENV"
}

setup_backend_env() {
  section "Backend: backend/.env"
  if [ -f "$BACKEND_ENV" ]; then
    skip_note "$BACKEND_ENV"
    return
  fi

  info "This is the Go API's configuration. Required values first, then a"
  info "series of optional integrations -- press Enter to skip any of those"
  info "and the matching feature degrades cleanly (the app still boots)."

  echo
  bold "-- Required --"
  ask PORT "API port" "4000"
  ask DATABASE_NAME "Postgres database name" "halendar"
  ask POSTGRES_USER "Postgres user" "halendar"
  ask_hidden POSTGRES_PASSWORD "Postgres password (blank to auto-generate)"
  if [ -z "$POSTGRES_PASSWORD" ]; then
    POSTGRES_PASSWORD="$(openssl rand -base64 18)"
    info "generated one for you"
  fi
  # docker-compose.yml maps Postgres to 127.0.0.1:5433 on the host -- the API
  # runs natively (not in Docker) for local dev, so DB_DSN_DEV needs the
  # host-side port, not Postgres' own internal 5432.
  DB_PORT="5433"
  ask FRONTEND_URL "Frontend URL (where OAuth connect flows redirect back to)" "http://localhost:9090"
  TRUSTED_ORIGINS="$FRONTEND_URL"
  ENV="development"

  info "Generating ENCRYPTION_KEY (encrypts connected email/calendar accounts at rest)..."
  ENCRYPTION_KEY="$(openssl rand -base64 32)"

  echo
  bold "-- Optional: outgoing mail (activation/welcome emails) --"
  info "Any SMTP account works. Leave blank to skip -- the app still runs,"
  info "it just won't send those emails."
  ask SMTP_HOST "SMTP host" ""
  if [ -n "$SMTP_HOST" ]; then
    ask SMTP_PORT "SMTP port" "587"
    ask SMTP_USERNAME "SMTP username" ""
    ask_hidden SMTP_PASSWORD "SMTP password"
    ask SMTP_SENDER "SMTP sender header" "Halendar <no-reply@yourdomain.com>"
  else
    SMTP_PORT="25"; SMTP_USERNAME=""; SMTP_PASSWORD=""; SMTP_SENDER=""
  fi

  echo
  bold "-- Optional: mailbox import (the core feature -- reads a real inbox for meeting requests) --"
  info "For Gmail: myaccount.google.com -> Security -> 2-Step Verification"
  info "must be on, then App passwords -> generate one for 'Mail'. Use that"
  info "16-character password below, not your normal Google password."
  info "Leave MAIL_USER blank to skip; /v1/mail/* will report mail as unconfigured."
  ask MAIL_USER "Mailbox address (e.g. you@gmail.com)" ""
  if [ -n "$MAIL_USER" ]; then
    ask MAIL_IMAP_HOST "IMAP host:port" "imap.gmail.com:993"
    ask MAIL_SMTP_HOST "SMTP host" "smtp.gmail.com"
    ask MAIL_SMTP_PORT "SMTP port" "587"
    ask_hidden MAIL_PASS "App password"
    ask MAIL_FROM "From header" "Halendar <${MAIL_USER}>"
    MAIL_IMAP_TLS="TRUE"; MAIL_FOLDER="INBOX"
  else
    MAIL_IMAP_HOST=""; MAIL_SMTP_HOST=""; MAIL_SMTP_PORT="587"; MAIL_PASS=""; MAIL_FROM=""
    MAIL_IMAP_TLS="TRUE"; MAIL_FOLDER="INBOX"
  fi

  echo
  bold "-- Optional: CalDAV calendar (e.g. iCloud), for availability checks --"
  info "For iCloud: appleid.apple.com -> Sign-In and Security ->"
  info "App-Specific Passwords -> generate one. Leave CALDAV_URL blank to skip."
  ask CALDAV_URL "CalDAV server URL" ""
  if [ -n "$CALDAV_URL" ]; then
    ask CALDAV_USER "CalDAV username/email" ""
    ask_hidden CALDAV_PASS "App-specific password"
    ask CALDAV_TIMEZONE "Timezone" "Europe/Paris"
  else
    CALDAV_USER=""; CALDAV_PASS=""; CALDAV_TIMEZONE="Europe/Paris"
  fi
  CALDAV_CALENDAR=""

  echo
  bold "-- Optional: push notifications (Firebase Cloud Messaging) --"
  info "Firebase Console (console.firebase.google.com) -> create/select a"
  info "project -> Project Settings -> Service Accounts -> Generate new"
  info "private key. Leave blank to skip; /v1/push/* reports itself as"
  info "unconfigured instead of failing."
  ask FCM_PROJECT_ID "Firebase project ID" ""
  if [ -n "$FCM_PROJECT_ID" ]; then
    local sa_path
    ask sa_path "Path to the downloaded service account JSON" ""
    if [ -n "$sa_path" ] && [ -f "$sa_path" ]; then
      cp "$sa_path" "$FIREBASE_SA"
      info "copied to $FIREBASE_SA"
    else
      info "no file copied -- place it at $FIREBASE_SA yourself before starting the API"
    fi
  fi
  # /app/... (the Dockerfile's WORKDIR) only applies to the deployed container
  # -- this .env is for running `go run ./cmd/api` natively from backend/.
  FCM_SERVICE_ACCOUNT_FILE="./firebase-service-account.json"

  echo
  bold "-- Optional: Google OAuth (Gmail import + Google Calendar connect) --"
  info "Google Cloud Console (console.cloud.google.com) -> APIs & Services ->"
  info "Credentials -> Create Credentials -> OAuth client ID -> Web application."
  info "Add BOTH redirect URLs below as authorized redirect URIs on that same"
  info "client. Leave blank to skip -- those connect buttons report"
  info "'unknown or unconfigured provider' instead of failing."
  ask GOOGLE_OAUTH_CLIENT_ID "OAuth client ID" ""
  if [ -n "$GOOGLE_OAUTH_CLIENT_ID" ]; then
    ask_hidden GOOGLE_OAUTH_CLIENT_SECRET "OAuth client secret"
    ask GOOGLE_OAUTH_REDIRECT_URL "Gmail connect redirect URL" "http://localhost:${PORT}/v1/email-accounts/gmail/callback"
    ask GOOGLE_OAUTH_CALENDAR_REDIRECT_URL "Calendar connect redirect URL" "http://localhost:${PORT}/v1/calendar-accounts/google/callback"
  else
    GOOGLE_OAUTH_CLIENT_SECRET=""; GOOGLE_OAUTH_REDIRECT_URL=""; GOOGLE_OAUTH_CALENDAR_REDIRECT_URL=""
  fi

  echo
  bold "-- AI classification --"
  info "gemma3:4b (default) is the tuned, reliable choice; gemma3:1b is"
  info "lighter/faster but produces more false positives in practice."
  ask OLLAMA_MODEL "Local Ollama model" "gemma3:4b"
  # The API runs natively for local dev (not inside docker-compose), so it
  # reaches Ollama via the host-mapped port, not the in-network hostname.
  OLLAMA_BASE_URL="http://localhost:11434"
  CLAUDE_MODEL="claude-haiku-4-5-20251001"
  GEMINI_MODEL="gemini-3.6-flash"
  EMAIL_SYNC_INTERVAL="1m"

  BACKEND_ENV_SUBST_KEYS="PORT ENV FRONTEND_URL TRUSTED_ORIGINS DATABASE_NAME POSTGRES_USER \
    POSTGRES_PASSWORD DB_PORT SMTP_HOST SMTP_PORT SMTP_USERNAME SMTP_PASSWORD \
    SMTP_SENDER MAIL_IMAP_HOST MAIL_IMAP_TLS MAIL_FOLDER MAIL_SMTP_HOST \
    MAIL_SMTP_PORT MAIL_USER MAIL_PASS MAIL_FROM CALDAV_URL CALDAV_USER \
    CALDAV_PASS CALDAV_CALENDAR CALDAV_TIMEZONE OLLAMA_MODEL OLLAMA_BASE_URL CLAUDE_MODEL \
    GEMINI_MODEL FCM_PROJECT_ID FCM_SERVICE_ACCOUNT_FILE ENCRYPTION_KEY \
    GOOGLE_OAUTH_CLIENT_ID GOOGLE_OAUTH_CLIENT_SECRET GOOGLE_OAUTH_REDIRECT_URL \
    GOOGLE_OAUTH_CALENDAR_REDIRECT_URL EMAIL_SYNC_INTERVAL"
  export $BACKEND_ENV_SUBST_KEYS
  render_backend_env
  info "wrote $BACKEND_ENV"
}

setup_frontend_env() {
  section "Frontend: frontend/lib/env.dart"
  if [ -f "$FRONTEND_ENV" ]; then
    skip_note "$FRONTEND_ENV"
    return
  fi

  info "Tells the Flutter app which backend to talk to in debug vs. release"
  info "builds (kDebugMode picks between them at runtime)."
  local backend_url frontend_url
  ask backend_url "Backend URL for local dev (debug builds)" "localhost:4355"
  ask frontend_url "Frontend URL for local dev (debug builds)" "localhost:9090"

  cat > "$FRONTEND_ENV" <<DART
import 'package:flutter/foundation.dart';

String backendURL = kDebugMode? "${backend_url}" : "your-production-backend.example.com";
String frontendURL = kDebugMode? "${frontend_url}" : "your-production-frontend.example.com";
DART
  info "wrote $FRONTEND_ENV (update the release-mode URLs before shipping a build)"
}

setup_google_services_json() {
  section "Frontend: Android push notifications (optional)"
  if [ -f "$GOOGLE_SERVICES" ]; then
    skip_note "$GOOGLE_SERVICES"
    return
  fi
  info "Only needed to build/run the Android app with push notifications"
  info "(frontend/lib/services/push_notifications.dart is Android-only)."
  info "Firebase Console -> Add app -> Android, package name from"
  info "frontend/android/app/build.gradle.kts (applicationId) -> download"
  info "google-services.json."
  local path
  ask path "Path to google-services.json (blank to skip)" ""
  if [ -n "$path" ] && [ -f "$path" ]; then
    cp "$path" "$GOOGLE_SERVICES"
    info "copied to $GOOGLE_SERVICES"
  else
    info "skipped -- Android push notifications won't work until this file is added"
  fi
}

bold "Halendar local dev setup"
info "Creates the config files needed to run Halendar locally. Anything"
info "that already exists is left alone, so this is safe to re-run."

setup_backend_env
setup_frontend_env
setup_google_services_json

section "Done"
info "Next: run 'make setup' from the repo root (if you haven't already) to"
info "bring up Postgres/Ollama and apply migrations, then 'make bootstrap-admin'"
info "to create your first login."

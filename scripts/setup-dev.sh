#!/usr/bin/env bash
# Brings up the local infrastructure Halendar's backend needs to run:
# creates the Docker volumes docker-compose.yml expects to already exist,
# seeds Postgres' config file into its volume before first boot, starts
# Postgres/Ollama/Adminer, and applies database migrations.
#
# Safe to re-run -- every step here is idempotent. Called by `make setup`
# after scripts/setup-env.sh; run directly with `bash scripts/setup-dev.sh`
# to redo just this part (e.g. after `docker volume rm`).
set -euo pipefail
cd "$(dirname "$0")/.."

bold() { printf '\033[1m%s\033[0m\n' "$*"; }
section() { echo; bold "== $* =="; }
info() { printf '  %s\n' "$*"; }

check_prereqs() {
  section "Checking prerequisites"
  local missing=0
  check() {
    if command -v "$1" >/dev/null 2>&1; then
      info "found $1"
    else
      echo "  MISSING: $1 -- $2"
      missing=1
    fi
  }
  check docker "https://docs.docker.com/get-docker/"
  check go "https://go.dev/doc/install"
  check flutter "https://docs.flutter.dev/get-started/install"
  check migrate "brew install golang-migrate/tap/migrate (macOS), or see https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation"
  if ! docker compose version >/dev/null 2>&1 && ! command -v docker-compose >/dev/null 2>&1; then
    echo "  MISSING: docker compose -- included with recent Docker Desktop installs"
    missing=1
  fi
  if [ "$missing" -eq 1 ]; then
    echo
    echo "Install the missing tools above, then re-run this script."
    exit 1
  fi
}

compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose -f backend/docker-compose.yml "$@"
  else
    docker-compose -f backend/docker-compose.yml "$@"
  fi
}

create_volumes() {
  section "Creating Docker volumes"
  # docker-compose.yml marks these "external: true" -- it expects them to
  # already exist rather than creating them itself. `docker volume create`
  # is a no-op if the name is already there, so this is safe to re-run.
  for vol in halendar-db halendar-api halendar-ollama config_files; do
    docker volume create "$vol" >/dev/null
    info "$vol"
  done
}

seed_postgres_config() {
  section "Seeding Postgres config"
  # postgres is started with `-c config_file=/etc/postgresql/postgresql.conf`
  # (see docker-compose.yml), which lives in the config_files volume above --
  # on a brand new volume that file doesn't exist yet, so Postgres would
  # refuse to start. Copy it in via a throwaway container before first boot.
  docker run --rm \
    -v config_files:/etc/postgresql \
    -v "$PWD/backend/postgresql.conf:/src/postgresql.conf:ro" \
    alpine cp /src/postgresql.conf /etc/postgresql/postgresql.conf
  info "copied backend/postgresql.conf into the config_files volume"
}

# backend/.env's DB_DSN_DEV uses ${VAR} interpolation that godotenv resolves
# at load time -- sourcing the file as plain bash (each line is a valid
# KEY="value" assignment) expands it the same way.
dev_dsn() {
  (set -a && source backend/.env >/dev/null 2>&1 && echo "$DB_DSN_DEV")
}

start_infra() {
  # The API writes data/halendar_log.log relative to its working directory --
  # in production the halendar-api volume supplies this, but for a natively
  # run API it just needs to exist (harmless to skip: it falls back to
  # logging on stdout instead).
  mkdir -p backend/data

  section "Starting Postgres, Ollama, and Adminer"
  # Not halendar-api: for local dev the API runs natively via `go run`, not
  # in this compose network (see scripts/setup-env.sh's OLLAMA_BASE_URL and
  # DB_DSN_DEV, both pointed at host-mapped ports for exactly this reason).
  compose up -d postgres ollama adminer
  info "waiting for Postgres to accept connections..."
  local pg_user
  pg_user=$(set -a && source backend/.env >/dev/null 2>&1 && echo "$POSTGRES_USER")
  for _ in $(seq 1 30); do
    if docker exec halendar-postgres pg_isready -U "$pg_user" >/dev/null 2>&1; then
      info "Postgres is up"
      return
    fi
    sleep 1
  done
  echo "Postgres did not become ready in time -- check 'docker logs halendar-postgres'." >&2
  exit 1
}

run_migrations() {
  section "Running database migrations"
  migrate -source file://backend/migrations -database "$(dev_dsn)" up
}

frontend_deps() {
  section "Installing frontend dependencies"
  (cd frontend && flutter pub get)
}

check_prereqs
create_volumes
seed_postgres_config
start_infra
run_migrations
frontend_deps

section "Local dev infrastructure is up"
info "Next:"
info "  1. make bootstrap-admin        # create your first login"
info "  2. (cd backend && go run ./cmd/api)"
info "  3. make -C frontend test       # runs the Flutter app in Chrome on :9090"

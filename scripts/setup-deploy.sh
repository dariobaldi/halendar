#!/usr/bin/env bash
# Interactively creates the config needed to deploy Halendar's backend to
# your own VPS: backend/remote/production/Caddyfile and api.service.
#
# Optional and independent of `scripts/setup-env.sh` / `make setup` -- only
# relevant once you actually have a server and a domain. Safe to re-run:
# existing files are left untouched (delete one first if you want to
# regenerate it). Run with `bash scripts/setup-deploy.sh`, or `make setup-deploy`.
set -euo pipefail
cd "$(dirname "$0")/.."

CADDYFILE="backend/remote/production/Caddyfile"
CADDYFILE_TEMPLATE="backend/remote/production/CaddyfileCONFIG"
API_SERVICE="backend/remote/production/api.service"
API_SERVICE_TEMPLATE="backend/remote/production/api.serviceCONFIG"
BACKEND_ENV="backend/.env"

bold() { printf '\033[1m%s\033[0m\n' "$*"; }
section() { echo; bold "== $* =="; }
info() { printf '  %s\n' "$*"; }

ask() {
  local __varname="$1" __question="$2" __default="${3-}" __answer
  if [ -n "$__default" ]; then
    read -r -p "  ${__question} [${__default}]: " __answer
  else
    read -r -p "  ${__question}: " __answer
  fi
  printf -v "$__varname" '%s' "${__answer:-$__default}"
}

bold "Halendar production deploy setup"
info "Only needed once you have your own VPS and domain to deploy to."
info "backend/remote/setup/01.sh (run via 'make -C backend deploy/setup')"
info "provisions a fresh Ubuntu server with Docker, Caddy, and the 'migrate'"
info "CLI -- run this script first so that provisioning step has a"
info "Caddyfile and api.service to install."

section "Server details"
ask DOMAIN "Domain that will point at your server (e.g. api.example.com)" ""
ask TLS_EMAIL "Contact email for Let's Encrypt TLS certificate notices" ""
ask HOST_IP "Server IP address" ""
ask DEPLOY_USER "Non-root user to deploy as (created by 01.sh if missing)" "halendar"

if [ -f "$CADDYFILE" ]; then
  info "$CADDYFILE already exists -- leaving it untouched."
else
  sed -e "s/your@email.com/${TLS_EMAIL}/" \
      -e "s/backend.domain.com/${DOMAIN}/" \
      -e "s/localhost:4000/localhost:4355/" \
      "$CADDYFILE_TEMPLATE" > "$CADDYFILE"
  info "wrote $CADDYFILE (reverse-proxies to :4355, the docker-compose"
  info "published port for halendar-api, not the app's internal PORT)"
fi

if [ -f "$API_SERVICE" ]; then
  info "$API_SERVICE already exists -- leaving it untouched."
else
  sed -e "s/\*\*user\*\*/${DEPLOY_USER}/g" -e "s/\*\*group\*\*/${DEPLOY_USER}/g" \
      "$API_SERVICE_TEMPLATE" > "$API_SERVICE"
  info "wrote $API_SERVICE"
fi

if [ -f "$BACKEND_ENV" ]; then
  section "Updating backend/.env"
  if grep -q '^HOST_IP=' "$BACKEND_ENV"; then
    sed -i.bak -e "s/^HOST_IP=.*/HOST_IP=${HOST_IP}/" -e "s/^USER=.*/USER=${DEPLOY_USER}/" "$BACKEND_ENV"
    rm -f "${BACKEND_ENV}.bak"
    info "set HOST_IP and USER (used by 'make -C backend connect' / 'deploy')"
  fi
else
  info "backend/.env doesn't exist yet -- run scripts/setup-env.sh (or 'make setup') first."
fi

section "Next steps"
info "1. Point a DNS A record for ${DOMAIN:-your domain} at ${HOST_IP:-your server IP}."
info "2. make -C backend deploy/setup   # provisions the server (ends in a reboot)"
info "3. make -C backend deploy         # applies migrations, syncs Caddy, deploys the API"

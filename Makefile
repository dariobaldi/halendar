# ============================================================================= #
# HELPERS	  																	#
# ============================================================================= #
.PHONY: help
help:
	@echo '*** Commands:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ============================================================================= #
# SETUP  																		#
# ============================================================================= #

## setup: first-time local dev setup -- creates config files, then brings up Postgres/Ollama and applies migrations
.PHONY: setup
setup:
	@bash scripts/setup-env.sh
	@bash scripts/setup-dev.sh

## setup-deploy: create/update production deploy config (Caddyfile, api.service) -- optional, re-run any time
.PHONY: setup-deploy
setup-deploy:
	@bash scripts/setup-deploy.sh

## bootstrap-admin: create the first user (full admin rights) on a fresh database
.PHONY: bootstrap-admin
bootstrap-admin:
	@$(MAKE) -C backend bootstrap-admin

# ============================================================================= #
# DEPLOYMENT  																	#
# ============================================================================= #

## deploy: deploy the backend, then the frontend (each prompts for its own confirmation/commit message)
.PHONY: deploy
deploy:
	@$(MAKE) -C backend deploy
	@$(MAKE) -C frontend deploy

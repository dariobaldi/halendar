# ============================================================================= #
# HELPERS	  																	#
# ============================================================================= #
.PHONY: help
help:
	@echo '*** Commands:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ============================================================================= #
# DEPLOYMENT  																	#
# ============================================================================= #

## deploy: deploy the backend, then the frontend (each prompts for its own confirmation/commit message)
.PHONY: deploy
deploy:
	@$(MAKE) -C backend deploy
	@$(MAKE) -C frontend deploy

# Parameters (e.g. make mails N=10)
#   N     number of mails
#   DAYS  days of calendar
#   FILE  JSON file (send, add)
N    ?= 5
DAYS ?= 7
FILE ?=

BIN := bin/halendar

.DEFAULT_GOAL := help
.PHONY: help build mails send calendar add clean

help: ## Show this help
	@printf "Usage: make <target> [PARAM=value]\n\n"
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'
	@printf "\nExamples:\n"
	@echo "  make mails N=10"
	@echo "  make send FILE=examples/send.json"
	@echo "  make add FILE=examples/event.json"

build: ## Compile the program
	@go build -o $(BIN) .

mails: build ## Recent mails             (N=5)
	@$(BIN) mails $(N)

send: build ## Sends a mail              (FILE=examples/send.json)
	@$(BIN) send $(or $(FILE),examples/send.json)

calendar: build ## Upcoming schedule     (DAYS=7)
	@$(BIN) calendar $(DAYS)

add: build ## Adds events               (FILE=examples/event.json)
	@$(BIN) add $(or $(FILE),examples/event.json)

clean: ## Removes the binary
	rm -rf bin

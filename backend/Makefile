include .env

# ============================================================================= #
# HELPERS	  																	#
# ============================================================================= #
.PHONY: help
help:
	@echo '*** Commands:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: ignored
ignored:
	@clear
	@echo "*** Ignored files ***\n"
	@git status --ignored --short | grep '!! ' | sed 's/!! //g'
	@echo

.PHONY: confirm
confirm:
	@echo "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

# ============================================================================= #
# DEVELOPMENT  																	#
# ============================================================================= #

## api/up: Update Docker image
.PHONY: api/up
api/up: build/api
	@docker context use default
	@echo '*** Building Docker Container'
	@docker-compose up -d --no-deps --build halendar-api

# ============================================================================= #
# MIGRATIONS 																	#
# ============================================================================= #

## migrate/create name=**** : Create a new database migration
.PHONY: migrate/create
migrate/create: confirm
	@migrate create -seq -ext sql -dir=./migrations ${name}

## migrate/up: Run database migrations
.PHONY: migrate/up
migrate/up: confirm
	@echo 'Running up migrations...'
	@migrate -source file://migrations -database ${DB_DSN_DEV} up

## migrate/fill: run the sql fill info file
.PHONY: migrate/fill
migrate/fill: migrate/up
	@PGPASSWORD=${POSTGRES_PASSWORD} psql -U ${POSTGRES_USER} -d ${DATABASE_NAME} -f ./sql_commands/hide_-fill_info.sql

## migrate/down: Rollback database migrations by 1 step
.PHONY: migrate/down
migrate/down: 
	@echo 'Running down migrations...'
	@migrate -source file://migrations -database ${DB_DSN_DEV} down

## migrate/goto version=####: Rollback database migrations to a specific version
.PHONY: migrate/goto
migrate/goto: confirm
	@echo 'Running up migrations...'
	@migrate -source file://migrations -database ${DB_DSN_DEV} goto ${version}

## migrate/force version=#### : Force database migrations to a specific version
.PHONY: migrate/force
migrate/force: confirm
	@echo 'Running up migrations...'
	@migrate -source file://migrations -database ${DB_DSN_DEV} force ${version}

# ============================================================================= #
# QUALITY CONTROL																#
# ============================================================================= #

## audit: Tidy dependencies and format, vet and test all code
.PHONY: audit
audit: tidy
	@echo '*** Formatting code'
	go fmt ./...
	@echo '*** Check vulnerabilities'
	govulncheck ./...
	@echo '*** Vetting code'
	staticcheck ./...
	@echo '*** Running tests'
	go test -race -vet=off ./...

## tidy: Tidy and vendor dependencies
.PHONY: tidy
tidy:
	@echo '*** Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	# @echo '*** Vendoring dependencies'
	# go mod vendor


# ============================================================================= #
# BUILD																			#
# ============================================================================= #

## build/api: Build the binary for the API server (Linux)
.PHONY: build/api
build/api:
	@echo '*** Building API binary'
	@timestamp=$$(date "+%Y-%m-%dT%H:%M:%S"); \
	GOOS=linux GOARCH=amd64 \
	go build -ldflags="-s -X main.buildTimestamp=$$timestamp" -o bin/linux_amd64/api ./cmd/api/

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

## connect: connect to the production server
.PHONY: connect
connect:
	@ssh ${USER}@${HOST_IP}

## connect/root: connect to the production server as root user
.PHONY: connect/root
connect/root:
	@ssh root@${HOST_IP}

## deploy/setup: install dependencies
.PHONY: deploy/setup
deploy/setup:
	@rsync -rP --delete ./remote/setup root@${HOST_IP}:/root
	@rsync -P .env root@${HOST_IP}:/root/setup/
	@ssh -t root@${HOST_IP} "bash /root/setup/01.sh"
	@rsync -P ./remote/production/Caddyfile ${USER}@${HOST_IP}:~
	@ssh -t ${USER}@${HOST_IP} '\
		sudo mv ~/Caddyfile /etc/caddy/ \
		&& sudo systemctl reload caddy \
	'

## deploy: deploy just the api executable
.PHONY: deploy
deploy: confirm
	@echo '*** Building API binary'
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o bin/linux_amd64/api ./cmd/api/
	@docker context use remote
	@echo '*** Building Docker Container'
	@docker-compose up -d --no-deps --build halendar-api
	@docker context use default

## deploy/restart: fully restart the docker container
.PHONY: deploy/restart
deploy/restart: confirm
	@docker context use remote
	@docker-compose down
	@docker-compose up -d
	@docker context use default

## db/backup: back-up production database
.PHONY: db/backup
db/backup:
	@echo '=> Creating db dump'
	@ssh ${USER}@${HOST_IP} rm -f ./db.dump
	@ssh ${USER}@${HOST_IP} docker exec postgres rm -f ./db.dump
	@ssh ${USER}@${HOST_IP} docker exec postgres pg_dump -U ${POSTGRES_USER} -d ${DATABASE_NAME} -F c -b -v -f ./db.dump
	@ssh ${USER}@${HOST_IP} docker cp postgres:/db.dump ./db.dump
	@echo '=> Importing db dump'
	@rm -f ./db.dump
	@scp ${USER}@${HOST_IP}:/home/${USER}/db.dump ./db.dump
	@echo '=> DB downloaded'


## db/import: import local db dump to postgres container
.PHONY: db/import
db/import: db/backup
	@echo '=> Importing db dump'
	@docker context use default
	@docker exec postgres rm -f ./db.dump
	@docker cp ./db.dump postgres:/db.dump
	@docker exec -i postgres pg_restore -c -U ${POSTGRES_USER} -d ${DATABASE_NAME} ./db.dump
	@echo '=> DB updated locally'
	@docker exec -i postgres psql -U ${POSTGRES_USER} -d ${DATABASE_NAME} -c "UPDATE accounts SET import_status=false AND import_products=false"

## db/config/local: update config file in local docker volume
.PHONY: db/config/local
db/config/local:
	@echo '=> Updating Config File locally'
	@docker cp ./postgresql.conf postgres:/etc/postgresql/postgresql.conf
	@docker-compose up -d --no-deps --build postgres

## db/config: update config file in docker volume
.PHONY: db/config
db/config:
	@echo '=> Updating Config File'
	@rsync -avz ./postgresql.conf halendar@51.89.150.207:~/postgresql.conf
	@ssh ${USER}@${HOST_IP} docker cp ./postgresql.conf postgres:/etc/postgresql/postgresql.conf
	@docker context use remote
	@docker-compose up -d --no-deps --build postgres
	@docker context use default

## frontend/version version=X.X.X: update production frontend version
.PHONY: frontend/version
frontend/version:
	@echo '=> Importing db dump'
	@docker context use remote
	@docker exec -i postgres psql -U ${POSTGRES_USER} -d ${DATABASE_NAME} -c "INSERT INTO app_version (frontend) VALUES ('${version}')"
	@docker exec -i postgres psql -U ${POSTGRES_USER} -d ${DATABASE_NAME} -c "REFRESH MATERIALIZED VIEW CONCURRENTLY dashboard_summary"
	@docker context use default

## apply: run new config
.PHONY: apply
apply:
	@echo '*** Building API binary'
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o bin/linux_amd64/api ./cmd/api/
	@docker context use remote
	@echo '*** Building Docker Container'
	@docker-compose up -d
	@docker context use default
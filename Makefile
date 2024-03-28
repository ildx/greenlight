include .envrc

# ============================== #
# HELPERS
# ============================== #

## help: print this help message
.PHONY: help
help:
	@echo "Usage:"
	@sed -n "s/^##//p" ${MAKEFILE_LIST} | column -t -s ":" | sed -e "s/^/ /"

.PHONY: confirm
confirm:
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]

# ============================== #
# DEVELOPMENT
# ============================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo "Starting the server..."
	go run ./cmd/api -db-dsn=$(DB_DSN)

## db/connect: connect to database using psql
.PHONY: db/connect
db/connect:
	@echo "Connecting to the database..."
	psql $(DB_DSN)

## migrate/create name=$1: create a new database migration
.PHONY: migrate/create
migrate/create:
	@echo "Creating $(name) migration..."
	migrate create -seq -ext=.sql -dir=./migration $(name)

## migrate/up: apply all database migrations
.PHONY: migrate/up
migrate/up: confirm
	@echo "Migrating..."
	migrate -source file://migrations -database $(DB_DSN) up

## migrate/down: rollback the last database migration
.PHONY: migrate/down
migrate/down: confirm
	@echo "Rolling back migrations..."
	migrate -source file://migrations -database $(DB_DSN) down

# ============================== #
# QUALITY CONTROL
# ============================== #

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo "Formatting code..."
	go fmt ./...
	@echo "Vetting code..."
	go vet ./...
	staticcheck ./...
	@echo "Running tests"
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo "Tidying and verifying module dependencies..."
	go mod tidy
	go mod verify
	@echo "Vendoring dependencies..."
	go mod vendor

# ============================== #
# BUILD
# ============================== #

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo "Building cmd/api..."
	go build -ldflags="-s" -o bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags="-s" -o=./bin/linux_amd64/api ./cmd/api

# ============================== #
# PRODUCTION
# ============================== #

production_host_ip = "XXX.XX.XX.XXX"

## production/connect: connect to the production server
.PHONY: production/connect
production/connect:
	@echo "Connecting to the production server..."
	ssh greenlight@$(production_host_ip)

## production/deploy/api: deploy api to production
.PHONY: production/deploy/api production/deploy/api:
	@echo "Deploying api to production..."
	rsync -P ./bin/linux_amd64/api greenlight@${production_host_ip}:~
	rsync -rP --delete ./migrations greenlight@${production_host_ip}:~
	rsync -P ./remote/production/api.service greenlight@${production_host_ip}:~ ssh -t greenlight@${production_host_ip} '\
	migrate -path ~/migrations -database $$GREENLIGHT_DB_DSN up \
		&& sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \

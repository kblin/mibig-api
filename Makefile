now := $(shell date -R)
sha := $(shell git rev-parse HEAD)

.PHONY: all serve test coverage

all:
	go build -ldflags "-X main.gitVer=$(sha) -X \"main.buildTime=$(now)\""

serve:
	go build -ldflags "-X main.gitVer=$(sha) -X \"main.buildTime=$(now)\""
	./mibig-api serve --debug

test:
	go test ./...

coverage:
	go test -coverprofile=cover.prof ./...
	go tool cover -html=cover.prof

.PHONY: tailwind-watch
tailwind-watch:
	tailwindcss -i ./static/css/input.css -o ./static/css/style.css --watch


.PHONY: templ-watch
templ-watch:
	templ generate -watch

.PHONY: templ-build
templ-build:
	templ generate

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path=./migrations -database ${MIBIG_DSN} up

.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations...'
	migrate -path=./migrations -database ${MIBIG_DSN} down

.PHONY: db/psql
db/psql:
	psql ${MIBIG_DSN}

.PHONY: integration
integration: all
	migrate -path=./migrations -database ${MIBIG_DSN} down -all
	migrate -path=./migrations -database ${MIBIG_DSN} up
	./integration.sh

.PHONY: local
local: all
	migrate -path=./migrations -database ${MIBIG_DSN} down -all
	migrate -path=./migrations -database ${MIBIG_DSN} up
	./load_externals.sh
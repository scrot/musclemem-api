SERVER_PACKAGE_PATH := ./cmd/server
SERVER_BINARY := api

OUTPUT_PATH := /tmp/musclemem-api/

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# tmp: creates directories for OUTPUT_PATH if needed
.PHONY: tmp
tmp: 
	mkdir -p ${OUTPUT_PATH}

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	gofmt -w .
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...


## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=${OUTPUT_PATH}coverage.out ./...
	go tool cover -html=${OUTPUT_PATH}coverage.out


## server/build: build the server
.PHONY: server/build
server/build: tmp tidy
	@go build -ldflags='-X main.version=$(shell git rev-parse --short HEAD)-snapshot' \
		-o=${OUTPUT_PATH}${SERVER_BINARY} ${SERVER_PACKAGE_PATH}

## server/run: run the server locally
.PHONY: server/run
server/run: server/build
	@ENVIRONMENT=development HOST=localhost PORT=8080 \
	${OUTPUT_PATH}${SERVER_BINARY}

## server/live: run the server with reloading on file changes
.PHONY: server/live
server/live:
	@ENVIRONMENT=development HOST=localhost PORT=8080 \
		go run github.com/cosmtrek/air@latest \
		--build.cmd "make server/kill && make server/build" \
		--build.bin "${OUTPUT_PATH}${SERVER_BINARY}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, sql" \
		--misc.clean_on_exit "true"

## server/kill: kills the server on running port :8080
.PHONY: server/kill
server/kill:
	lsof -t -i:8080 | xargs -r kill

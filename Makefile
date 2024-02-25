SERVER_PACKAGE_PATH := ./cmd/server
APP_NAME := musclemem-api

OUTPUT_PATH := /tmp/musclemem-api/

PROJECT_ID := musclemem
GC_REGION := europe-west1

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

## kill: kills the server on running port :8080
.PHONY: kill
kill:
	lsof -t -i:8080 | xargs -r kill

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=${OUTPUT_PATH}coverage.out ./...
	go tool cover -html=${OUTPUT_PATH}coverage.out


## build: build the server
.PHONY: build
build: tmp tidy
	@go build -ldflags='-X main.version=$(shell git rev-parse --short HEAD)-snapshot' \
		-o=${OUTPUT_PATH}${APP_NAME} ${SERVER_PACKAGE_PATH}

## run: run the server locally
.PHONY: run
run: build
	@ENVIRONMENT=development HOST=localhost PORT=8080 \
	${OUTPUT_PATH}${APP_NAME}

.PHONY: run/prod
run/prod: build
	@ENVIRONMENT=production HOST=localhost PORT=8080 DATABASE_DSN=${MUSCLEMEM_DB_DSN} \
	${OUTPUT_PATH}${APP_NAME}

## live: run the server with reloading on file changes
.PHONY: live
live:
	@ENVIRONMENT=development HOST=localhost PORT=8080 \
		go run github.com/cosmtrek/air@latest \
		--build.cmd "make kill && make build" \
		--build.bin "${OUTPUT_PATH}${APP_NAME}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, sql" \
		--misc.clean_on_exit "true"


## deploy: deploy to google cloud run
.PHONY: deploy
deploy:
	@gcloud builds submit --tag gcr.io/${PROJECT_ID}/${APP_NAME} .
	@gcloud run deploy ${APP_NAME} --image gcr.io/${PROJECT_ID}/${APP_NAME} --region ${GC_REGION}

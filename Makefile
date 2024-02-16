SERVER_PACKAGE_PATH := ./cmd/server
SERVER_BINARY := musclemem-api

CLI_PACKAGE_PATH := ./cmd/cli
CLI_BINARY := mm

OUTPUT_PATH := /tmp/${BINARY_NAME}

GITHUB_UNAME := scrot
# GITHUB_TOKEN export this variable

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

.PHONY: no-dirty
no-dirty:
	git diff --exit-code


# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #


## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	gofumpt -w .
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

# ==================================================================================== #
# SERVER
# ==================================================================================== #

## server/build: build the server
.PHONY: server/build
server/build: tidy
	@go build -ldflags='-X main.version=$(shell git rev-parse --short HEAD)-snapshot' \
		-o=${OUTPUT_PATH}${SERVER_BINARY} ${SERVER_PACKAGE_PATH}

## server/run: run the server locally
.PHONY: server/run
server/run: server/build
	@${OUTPUT_PATH}${SERVER_BINARY}

## server/live: run the server with reloading on file changes
.PHONY: server/live
server/live:
	@go run github.com/cosmtrek/air@latest \
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

# ==================================================================================== #
# CLI
# ==================================================================================== #

BVAR := cmd/cli/main
## cli/build: build the server
.PHONY: cli/build
cli/build: tidy
	@go build \
	  -ldflags="-X '${BVAR}.name=${CLI_BINARY}' -X '${BVAR}.version=$(shell git rev-parse --short HEAD)-snapshot' -X '${BVAR}.date=$(shell date)'"\
		-o=${OUTPUT_PATH}${CLI_BINARY} ${CLI_PACKAGE_PATH}

## cli/run: run the server locally
.PHONY: cli/run
cli/run: cli/build
	@${OUTPUT_PATH}${CLI_BINARY}

## cli/init: loads the testdata
.PHONY: cli/init 
cli/init: cli/build
	${OUTPUT_PATH}${CLI_BINARY} logout
	${OUTPUT_PATH}${CLI_BINARY} register -f testdata/user.json
	${OUTPUT_PATH}${CLI_BINARY} login \
		--username $(shell jq -r '.username' testdata/user.json) \
		--password $(shell jq -r '.password' testdata/user.json)
	${OUTPUT_PATH}${CLI_BINARY} add wo -f testdata/workout.json
	${OUTPUT_PATH}${CLI_BINARY} add wo -f testdata/workouts.json
	${OUTPUT_PATH}${CLI_BINARY} add ex 1 -f testdata/exercise.json
	${OUTPUT_PATH}${CLI_BINARY} add ex 1 -f testdata/exercises.json
	${OUTPUT_PATH}${CLI_BINARY} move ex down 1/1
	${OUTPUT_PATH}${CLI_BINARY} move ex up 1/2
	${OUTPUT_PATH}${CLI_BINARY} move ex swap 1/1 1/2
	${OUTPUT_PATH}${CLI_BINARY} edit ex 1/1 --name "CHANGED" --weight 999.9 --reps 9
	${OUTPUT_PATH}${CLI_BINARY} edit wo 1 --name "CHANGED"
	${OUTPUT_PATH}${CLI_BINARY} remove ex 1/2
	${OUTPUT_PATH}${CLI_BINARY} remove wo 2
	${OUTPUT_PATH}${CLI_BINARY} list ex 1
	${OUTPUT_PATH}${CLI_BINARY} list wo


## cli/alias: creates a temporary alias to cli command
.PHONY: cli/alias
cli/alias:
	alias mm='make cli/build && ${OUTPUT_PATH}${CLI_BINARY}'

# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

# run/docker: create and run docker image in docker environment
.PHONY: run/docker
run/docker: build
	docker build -t ${SERVER_BINARY} -f Dockerfile.goreleaser .	
	docker run --rm -p 8080:80 ${SERVER_BINARY}

# run/kubernetes: create and run project in k8s environment
.PHONY: run/kubernetes
run/kubernetes: build
	kubectl delete -f ./charts/musclemem-api.yaml
	kubectl apply -f ./charts/musclemem-api.yaml
	kubectl port-forward deployment/musclemem-api 8080:80

## kube/ghcr-creds: store ghcr.io container regestry creds in kubernetes
.PHONY: kube/ghcr-creds
kube/creds:
	kubectl create secret docker-registry regcred --docker-server=ghcr.io --docker-username=${GITHUB_UNAME} --docker-password=${GITHUB_TOKEN}

## push: push changes to the remote Git repository
.PHONY: push
push: tidy audit no-dirty
	git push

.PHONY: release
release:
	GITHUB_TOKEN=${GITHUB_TOKEN} goreleaser release --clean

## production/deploy: deploy the application to production
# .PHONY: production/deploy
# production/deploy: confirm tidy audit no-dirty
# 	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=/tmp/bin/linux_amd64/${BINARY_NAME} ${MAIN_PACKAGE_PATH}
# 	upx -5 /tmp/bin/linux_amd64/${BINARY_NAME}

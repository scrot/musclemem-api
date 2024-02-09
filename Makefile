SERVER_PACKAGE_PATH := ./cmd/server
SERVER_BINARY := musclemem-api

CLI_PACKAGE_PATH := ./cmd/cli
CLI_BINARY := musclemem-cli

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

## kill: kills process on running port :8080
.PHONY: kill
kill:
	lsof -t -i:8080 | xargs -r kill

## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v

## audit: run quality control checks
.PHONY: audit
audit:
	go mod verify
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	go test -race -buildvcs -vet=off ./...


# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## test: run all tests
.PHONY: test
test:
	go test -v -race -buildvcs ./...

## test/cover: run all tests and display coverage
.PHONY: test/cover
test/cover:
	go test -v -race -buildvcs -coverprofile=${OUTPUT_PATH}/coverage.out ./...
	go tool cover -html=${OUTPUT_PATH}/coverage.out

## build: build the application
.PHONY: build/server
build/server: tidy
	@go build -ldflags='-X main.version=$(shell git rev-parse --short HEAD)-snapshot' \
		-o=${OUTPUT_PATH}/${SERVER_BINARY} ${SERVER_PACKAGE_PATH}

## run: run the application
.PHONY: run/server
run/server: build/server
	@${OUTPUT_PATH}/${SERVER_BINARY}

## run/live: run the application with reloading on file changes
.PHONY: live/server
live/server:
	@go run github.com/cosmtrek/air@latest \
		--build.cmd "make kill && make build/server" \
		--build.bin "${OUTPUT_PATH}/${SERVER_BINARY}" \
		--build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, sql" \
		--misc.clean_on_exit "true"

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
	

# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

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

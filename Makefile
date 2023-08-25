# Change these variables as necessary.
MAIN_PACKAGE_PATH := ./cmd/api
BINARY_NAME := musclemem-api
GITHUB_UNAME := scrot

# Export the following environment variables 
# GITHUB_TOKEN

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
	go test -v -race -buildvcs -coverprofile=/tmp/coverage.out ./...
	go tool cover -html=/tmp/coverage.out

## build: build the application
.PHONY: build
build:
	GITHUB_TOKEN=${GITHUB_TOKEN} goreleaser build --snapshot --clean
# go build -o=/tmp/bin/${BINARY_NAME} ${MAIN_PACKAGE_PATH}

## run: run the application
.PHONY: run
run: build
	/tmp/bin/${BINARY_NAME}

## run/live: run the application with reloading on file changes
.PHONY: run/live
run/live:
	go run github.com/cosmtrek/air@v1.43.0 \
		--build.cmd "make build" --build.bin "/tmp/bin/${BINARY_NAME}" --build.delay "100" \
		--build.exclude_dir "" \
		--build.include_ext "go, tpl, tmpl, html, css, scss, js, ts, sql, jpeg, jpg, gif, png, bmp, svg, webp, ico" \
		--misc.clean_on_exit "true"

# run/docker: create and run docker image in docker environment
.PHONY: run/docker
run/docker: build
	docker build -t ${BINARY_NAME} -f Dockerfile.goreleaser .	
	docker run --rm -p 8080:80 ${BINARY_NAME}

# run/kubernetes: create and run project in k8s environment
.PHONY: run/kubernetes
run/kubernetes: build
	kubectl delete -f ./charts/musclemem-api.yaml
	kubectl apply -f ./charts/musclemem-api.yaml
	

# ==================================================================================== #
# OPERATIONS
# ==================================================================================== #

## kube/creds: store regcred for project images in kubernetes
.PHONY: kube/creds
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
.PHONY: production/deploy
production/deploy: confirm tidy audit no-dirty
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=/tmp/bin/linux_amd64/${BINARY_NAME} ${MAIN_PACKAGE_PATH}
	upx -5 /tmp/bin/linux_amd64/${BINARY_NAME}

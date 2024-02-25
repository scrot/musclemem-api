# syntax=docker/dockerfile:1
# https://docs.docker.com/language/golang/build-images/
# https://snyk.io/blog/containerizing-go-applications-with-docker/

# Builder image
FROM golang:1.22.0-alpine3.19 AS builder

WORKDIR /usr/src/app

COPY . ./
RUN go mod download &&\
  go build -o=/tmp/musclemem-api ./cmd/server

# Production image
FROM  gcr.io/distroless/base:latest

LABEL maintainer="Roy de Wildt"

COPY --from=builder /tmp/musclemem-api /usr/local/bin/

EXPOSE 8080

CMD ["musclemem-api"]

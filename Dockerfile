# syntax=docker/dockerfile:1
# https://docs.docker.com/language/golang/build-images/
# https://snyk.io/blog/containerizing-go-applications-with-docker/

# Builder image
FROM golang:1.22.0-alpine3.19 AS builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=arm64

RUN addgroup --system nonroot && \ 
  adduser --system nonroot --ingroup nonroot

WORKDIR /app

COPY . ./
RUN go mod download &&\
  go mod verify &&\
  go build -o=/app/musclemem-api ./cmd/server

# Production image
FROM scratch
LABEL maintainer="Roy de Wildt"

COPY --from=builder /app/musclemem-api /
COPY --from=builder /app/musclemem.sqlite /
COPY --from=builder /etc/passwd /etc/passwd

USER nonroot

ENV PORT=8080
EXPOSE 8080

CMD ["/musclemem-api"]

# syntax=docker/dockerfile:1
FROM --platform=arm64 golang:latest AS build

WORKDIR /src/

COPY . .

RUN go mod download
RUN go mod verify

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=arm64

RUN go build -o /out/mm-api ./cmd/api

FROM --platform=arm64 gcr.io/distroless/static-debian11

LABEL maintainer="Roy de Wildt"

COPY --from=build /out/mm-api .

USER nonroot:nonroot

ENV PORT=80

EXPOSE 80

ENTRYPOINT [ "./mm-api" ]

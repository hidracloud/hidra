ARG GOLANG_VERSION=1.19.1
ARG ALPINE_VERSION=3.16

FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download 

COPY . .

RUN go build -o hidra .

FROM golang:${GOLANG_VERSION}-buster as release-maker

FROM alpine:${ALPINE_VERSION} AS production

RUN apk add --no-cache ca-certificates chromium-chromedriver

COPY --from=builder /app/hidra /usr/local/bin/hidra

# RUN addgroup -S hidra && adduser -S -G hidra hidra
# USER hidra
USER root

ENTRYPOINT ["/usr/local/bin/hidra"]
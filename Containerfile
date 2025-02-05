ARG GO_VERSION=1.23.6-alpine3.21
ARG DELVE_VERSION=1.24.0
ARG ALPINE_VERSION=3.21.2


#### Base image ####
FROM golang:${GO_VERSION} AS base

WORKDIR /go/src/app

ENV CGO_ENABLED=0
ENV PROMPT="%B%F{cyan}%n%f@%m:%F{yellow}%~%f %F{%(?.green.red[%?] )}>%f %b"

ARG DELVE_VERSION

RUN apk add --no-cache \
        ca-certificates \
        chromium-chromedriver \
        git \
        zsh \
 && go install github.com/go-delve/delve/cmd/dlv@v${DELVE_VERSION}

ARG USER_ID=1000
ENV USER_NAME=default

RUN adduser -D -u ${USER_ID} ${USER_NAME}

RUN chown -R ${USER_NAME}: /go

USER ${USER_NAME}


#### Builder image ####
FROM base AS builder

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o hidra main.go


#### Runtime image ####
FROM alpine:${ALPINE_VERSION} AS runtime

WORKDIR /usr/local/bin

RUN apk add --no-cache \
        ca-certificates \
        chromium-chromedriver

RUN adduser -D hidra

COPY --from=builder /go/src/app/hidra .

USER hidra

CMD ["./hidra"]

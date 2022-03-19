FROM alpine:3.14 as base
RUN apk add --no-cache\
    ca-certificates\
    chromium-chromedriver

FROM golang:1.18-alpine as build

WORKDIR /app
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hidra src/cmd/hidra/main.go

FROM base as runtime

COPY --from=build /app/hidra /usr/local/bin/hidra

ENTRYPOINT [ "/usr/local/bin/hidra" ]

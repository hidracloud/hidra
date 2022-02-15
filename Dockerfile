FROM alpine:3.14 as base
RUN apk add --no-cache\
    ca-certificates\
    chromium-chromedriver


FROM golang:1.17-alpine as build

# RUN apt update && apt install build-essential git -y

WORKDIR /app
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build  -a -installsuffix cgo -o hidra cmd/hidra/main.go

FROM base as runtime

COPY --from=build /app/hidra /usr/local/bin/hidra

ENTRYPOINT [ "/usr/local/bin/hidra" ]

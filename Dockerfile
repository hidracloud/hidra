FROM golang:1.16-alpine as build

WORKDIR /app
COPY . .

RUN go mod vendor
RUN  CGO_ENABLED=0 GOOS=linux go build -o hidra cmd/hidra/main.go

FROM alpine:3 as runtime

ARG DATA_DIR="/var/lib/hidra/data"
RUN mkdir -p $DATA_DIR

COPY --from=build /app/hidra /usr/local/bin/hidra

ENTRYPOINT [ "/usr/local/bin/hidra" ]
FROM golang:1.16-bullseye as build

RUN apt update && apt install build-essential git -y

WORKDIR /app
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

RUN VERSION=$(git describe --tags --abbrev=0) && \
    sed -i "s/latest/${VERSION}/g" utils/version.go
RUN CGO_ENABLED=1 GOOS=linux go build -o hidra cmd/hidra/main.go

FROM chromedp/headless-shell:stable as runtime

ARG DATA_DIR="/var/lib/hidra/data"
RUN mkdir -p $DATA_DIR

# Install ca-certificates for debian-based systems
RUN apt-get update && apt-get install -y ca-certificates libc6
RUN apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=build /app/hidra /usr/local/bin/hidra

ENTRYPOINT [ "/usr/local/bin/hidra" ]

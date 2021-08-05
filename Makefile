BUILD_PATH = 'build'
VERSION = $(shell git describe --tags --abbrev=0)

build:
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-amd64/hidra cmd/hidra/main.go
	GOOS=linux GOARCH=386 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-386/hidra cmd/hidra/main.go
	GOOS=linux GOARCH=arm64 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-arm64/hidra cmd/hidra/main.go
	GOOS=linux GOARCH=arm go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-arm/hidra cmd/hidra/main.go
	GOOS=darwin GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-amd64/hidra cmd/hidra/main.go
	GOOS=darwin GOARCH=arm64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-arm64/hidra cmd/hidra/main.go
deps:
	go mod vendor

compress:
	cd ${BUILD_PATH} && \
	tar -czf hidra-${VERSION}-linux-amd64.tar.gz hidra-${VERSION}-linux-amd64 && \
	tar -czf hidra-${VERSION}-linux-386.tar.gz hidra-${VERSION}-linux-386 && \
	tar -czf hidra-${VERSION}-linux-arm.tar.gz hidra-${VERSION}-linux-arm && \
	tar -czf hidra-${VERSION}-linux-arm64.tar.gz hidra-${VERSION}-linux-arm64 && \
	tar -czf hidra-${VERSION}-darwin-amd64.tar.gz hidra-${VERSION}-darwin-amd64 && \
	tar -czf hidra-${VERSION}-darwin-arm64.tar.gz hidra-${VERSION}-darwin-arm64

clean:
	rm -rf build

all: deps build compress
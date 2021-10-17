BUILD_PATH = 'build'
VERSION = $(shell git describe --tags --abbrev=0)

versionize:
	sed -i 's/latest/${VERSION}/g' utils/version.go

build:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-amd64/hidra cmd/hidra/main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-386/hidra cmd/hidra/main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-arm64/hidra cmd/hidra/main.go
	CGO_ENABLED=1 GOOS=linux GOARCH=arm go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-arm/hidra cmd/hidra/main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-amd64/hidra cmd/hidra/main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-arm64/hidra cmd/hidra/main.go
deps:
	go mod vendor

compress:
	cd ${BUILD_PATH} && \
	tar -czf hidra-${VERSION}-linux-amd64.tar.gz hidra-${VERSION}-linux-amd64 && \
	tar -czf hidra-${VERSION}-linux-arm64.tar.gz hidra-${VERSION}-linux-arm64 && \
	tar -czf hidra-${VERSION}-darwin-amd64.tar.gz hidra-${VERSION}-darwin-amd64 && \
	tar -czf hidra-${VERSION}-darwin-arm64.tar.gz hidra-${VERSION}-darwin-arm64

clean:
	rm -rf build
	sed -i 's/${VERSION}/latest/g' utils/version.go


all: deps build compress

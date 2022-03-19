BUILD_PATH = 'build'
VERSION = $(shell git describe --tags --abbrev=0)

versionize:
	sed -i 's/latest/${VERSION}/g' src/utils/version.go

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-amd64/hidra src/cmd/hidra/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-386/hidra src/cmd/hidra/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-arm64/hidra src/cmd/hidra/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-arm/hidra src/cmd/hidra/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-amd64/hidra src/cmd/hidra/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-arm64/hidra src/cmd/hidra/main.go
	bash deb/DEBIAN/control.sh amd64
	cp ${BUILD_PATH}/hidra-${VERSION}-linux-amd64/hidra deb/usr/local/bin/hidra
	dpkg-deb --build deb build/hidra-${VERSION}-amd64.deb
	bash deb/DEBIAN/control.sh arm64
	cp ${BUILD_PATH}/hidra-${VERSION}-linux-arm64/hidra deb/usr/local/bin/hidra
	dpkg-deb --build deb build/hidra-${VERSION}-arm64.deb
deps:
	go mod vendor

tests:
	go test -v ./...
	
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
	sed -i 's/${VERSION}/latest/g' src/utils/version.go


all: deps build compress

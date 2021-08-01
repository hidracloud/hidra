BUILD_PATH = 'build'
VERSION = v1.0.0-alpha.1

build:
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-amd64/hidra cmd/hidra/main.go
	GOOS=linux GOARCH=386 go build -o ${BUILD_PATH}/hidra-${VERSION}-linux-386/hidra cmd/hidra/main.go
	GOOS=darwin GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-amd64/hidra cmd/hidra/main.go
	GOOS=darwin GOARCH=arm64 go build -o ${BUILD_PATH}/hidra-${VERSION}-darwin-386/hidra cmd/hidra/main.go
	GOOS=windows GOARCH=amd64 go build -o ${BUILD_PATH}/hidra-${VERSION}-windows-amd64/hidra.exe cmd/hidra/main.go
	GOOS=windows GOARCH=386 go build -o ${BUILD_PATH}/hidra-${VERSION}-windows-386/hidra.exe cmd/hidra/main.go	
deps:
	go mod vendor

compress:
	cd ${BUILD_PATH} && \
	tar -czf hidra-${VERSION}-linux-amd64.tar.gz hidra-${VERSION}-linux-amd64 && \
	tar -czf hidra-${VERSION}-linux-386.tar.gz hidra-${VERSION}-linux-386 && \
	tar -czf hidra-${VERSION}-darwin-amd64.tar.gz hidra-${VERSION}-darwin-amd64 && \
	tar -czf hidra-${VERSION}-darwin-386.tar.gz hidra-${VERSION}-darwin-386 && \
	tar -czf hidra-${VERSION}-windows-amd64.tar.gz hidra-${VERSION}-windows-amd64 && \
	tar -czf hidra-${VERSION}-windows-386.tar.gz hidra-${VERSION}-windows-386 

clean:
	rm -rf build

all: deps build compress
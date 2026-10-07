APP := raw-img-cleanup
DIST := dist
HOST_OS := $(shell go env GOOS)
HOST_ARCH := $(shell go env GOARCH)
HOST_EXT := $(if $(filter windows,$(HOST_OS)),.exe)

.PHONY: all build test build-all build-mac build-win build-linux clean

all: build

build:
	mkdir -p $(DIST)
	go build -o $(DIST)/$(APP)-$(HOST_OS)-$(HOST_ARCH)$(HOST_EXT) .

test:
	go test ./...

build-all: build-mac build-win build-linux

build-mac:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(DIST)/$(APP)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(DIST)/$(APP)-darwin-arm64 .

build-win:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(DIST)/$(APP)-windows-amd64.exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o $(DIST)/$(APP)-windows-arm64.exe .

build-linux:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(DIST)/$(APP)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(DIST)/$(APP)-linux-arm64 .

clean:
	rm -rf $(DIST)

APP := rawtidy
DIST := dist
HOST_OS := $(shell go env GOOS)
HOST_ARCH := $(shell go env GOARCH)
HOST_EXT := $(if $(filter windows,$(HOST_OS)),.exe)

.PHONY: all build test build-all build-mac build-win build-linux clean

all: build

build:
	mkdir -p $(DIST)
	go build -o $(DIST)/$(APP)$(HOST_EXT) .

test:
	go test ./...

build-all: build-mac build-win build-linux

build-mac:
	mkdir -p $(DIST)/darwin-amd64 $(DIST)/darwin-arm64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(DIST)/darwin-amd64/$(APP) .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $(DIST)/darwin-arm64/$(APP) .

build-win:
	mkdir -p $(DIST)/windows-amd64 $(DIST)/windows-arm64
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(DIST)/windows-amd64/$(APP).exe .
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o $(DIST)/windows-arm64/$(APP).exe .

build-linux:
	mkdir -p $(DIST)/linux-amd64 $(DIST)/linux-arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(DIST)/linux-amd64/$(APP) .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(DIST)/linux-arm64/$(APP) .

clean:
	rm -rf $(DIST)

.PHONY: all

BINARY=sapmon
BUILD_VERSION=`git describe --tags`
BUILD_DATE=`date +%FT%T%z`

all: clean build_windows build_linux

build: build_windows build_linux

build_windows:
	GOOS=windows go build  -ldflags "-s -w -X main.BUILD_VERSION=$(BUILD_VERSION) -X main.BUILD_DATE=$(BUILD_DATE)" -o $(BINARY).exe

build_linux:
	GOOS=linux go build  -ldflags "-s -w -X main.BUILD_VERSION=$(BUILD_VERSION) -X main.BUILD_DATE=$(BUILD_DATE)" -o $(BINARY)

clean:
	rm -f $(BINARY)
	rm -f $(BINARY).exe

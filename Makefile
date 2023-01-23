# Makefile for the building of the daemon and production of a PSF package for installation on PLDist

TAG_COMMIT := $(shell git rev-list --abbrev-commit --tags --max-count=1)
TAG := $(shell git describe --abbrev=0 --tags ${TAG_COMMIT} 2>/dev/null || true)
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
VERSION := $(TAG:v%=%)
ifneq ($(COMMIT), $(TAG_COMMIT))
	VERSION := $(VERSION)-next-$(COMMIT)-$(DATE)
endif
ifeq ($(VERSION),)
	VERSION := $(COMMIT)-$(DATA)
endif
ifneq ($(shell git status --porcelain),)
	VERSION := $(VERSION)-dirty
endif

# Linker flags for go build
FLAGS := -ldflags "-X main.version=$(VERSION)"

default: clean linux-amd64 linux-arm64 windows-amd64

clean:
	rm -rf ./bin

linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(FLAGS) -o bin/evedbtool .

linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(FLAGS) -o bin/evedb_aarch64 .

windows-amd64:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(FLAGS) -o bin/evedbtool.exe .

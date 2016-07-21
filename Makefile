# Get the short git sha.
GIT_SHA := $(shell git log -1 --pretty=format:"%h")
# Get the first entry from GOPATH. This is an issue because the unusual
# configuration on CircleCI ends up with multiple GOPATHs, but it shouldn't
# be a problem elsewhere.
ONEGOPATH := $(firstword $(subst :, ,$(GOPATH)))
# Default the build number to 0, if its not defined.
BUILD_NUM ?= 0

all: install

install:
	glide install

test:
	go test -cover $(glide nv)

build:
	go build \
		-ldflags "-X main.progBuild=$(GIT_SHA) -X main.progReleaseNum=$(BUILD_NUM)" \
		-o $(ONEGOPATH)/bin/infomodels 

dist: dist-build dist-zip

dist-build:
	mkdir -p dist

	gox -output="dist/{{.OS}}-{{.Arch}}/infomodels" \
		-ldflags "-X main.progBuild=$(GIT_SHA) -X main.progReleaseNum=$(BUILD_NUM)" \
		-os="linux windows darwin" \
		-arch="amd64" > /dev/null

dist-zip:
	cd dist && zip infomodels-linux-amd64.zip linux-amd64/*
	cd dist && zip infomodels-windows-amd64.zip windows-amd64/*
	cd dist && zip infomodels-darwin-amd64.zip darwin-amd64/*


.PHONY: test dist

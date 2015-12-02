default: gen
VERSION := $(shell git describe --always --dirty --tags)
TOOLS := github.com/jteeuwen/go-bindata/go-bindata github.com/tinylib/msgp github.com/mitchellh/gox github.com/aktau/github-release

deps:
	glide -q -y glide.yaml up

tools:
	go get ${TOOLS}

tools-update:
	go get -u -f ${TOOLS}

clean:
	rm -f cmd/pipeviz/pipeviz cmd/pipeviz/pipeviz.test
	rm -f cmd/pvutil/pvutil cmd/pvutil/pvutil.test
	rm -f cmd/pvproxy/pvproxy cmd/pvproxy/pvproxy.test

gen: tools
	go generate -x ./schema
	go-bindata -o fixtures/bindata_fixtures.go -prefix="fixtures" -pkg=fixtures fixtures/*/*.json

test: gen deps
	go test $(shell glide nv)

install: gen deps
	go install -ldflags "-X github.com/pipeviz/pipeviz/version.v=${VERSION}" ./cmd/...

build-all: gen deps
	gox -verbose \
	-ldflags "-X github.com/pipeviz/pipeviz/version.v=${VERSION}" \
	-os="linux darwin freebsd" \
	-arch="amd64 386" \
	-output="dist/{{.OS}}-{{.Arch}}/{{.Dir}}" ./cmd/...

.PHONY: deps tools tools-update gen install clean build-all

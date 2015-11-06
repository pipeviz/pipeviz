default: gen
VERSION := $(shell git describe --always --dirty --tags)

deps:
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u github.com/tinylib/msgp
	go get -u github.com/mitchellh/gox

clean:
	rm -f cmd/pipeviz/pipeviz cmd/pipeviz/pipeviz.test
	rm -f cmd/pvutil/pvutil cmd/pvutil/pvutil.test
	rm -f cmd/pvproxy/pvproxy cmd/pvproxy/pvproxy.test

test:
	go test ./...

gen: deps
	go generate -x ./schema

install:
	go install -ldflags "-X main.version=${VERSION}" ./cmd/...

.PHONY: deps gen install clean

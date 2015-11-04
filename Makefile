default: gen

deps:
	go get -d ./...
	go get github.com/jteeuwen/go-bindata/...
	go get github.com/tinylib/msgp

gen: deps
	go generate -x ./schema

install:
	go install
	go install ./cmd/...

client-deps:
	go get -d ./clients/...
	go install ./clients/cmd/...

install-all: install
	go install ./clients/cmd/...

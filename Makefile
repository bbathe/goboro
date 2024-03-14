package := $(shell basename `pwd`)

.PHONY: default get codetest build setup fmt lint vet

default: fmt codetest

get:
ifneq ("$(CI)", "true")
	go get -u ./...
	go mod tidy
endif
	go mod download
	go mod verify

codetest: lint vet

build: default
	mkdir -p target
	rm -f target/$(package).exe target/$(package).log
	go get github.com/akavel/rsrc
	go install github.com/akavel/rsrc
	$(shell go env GOPATH)/bin/rsrc -arch amd64 -manifest $(package).manifest -ico $(package).ico -o cmd/goboro/$(package).syso
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC="x86_64-w64-mingw32-gcc" go build -v -ldflags "-s -w -H=windowsgui" -o target/$(package).exe github.com/bbathe/goboro/cmd/goboro
	go mod tidy

fmt:
	GOOS=windows GOARCH=amd64 go fmt ./...

lint:
ifeq ("$(CI)", "true")
	GOOS=windows GOARCH=amd64 $(shell go env GOPATH)/bin/golangci-lint run --verbose --timeout 3m
else
	GOOS=windows GOARCH=amd64 $(shell go env GOPATH)/bin/golangci-lint run --fix
endif

vet:
	GOOS=windows GOARCH=amd64 go vet -all ./...

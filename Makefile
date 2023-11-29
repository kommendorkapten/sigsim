all: build

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: bulid
build:
	go build -ldflags="-s -w" -trimpath -o sigsim ./cmd/sigsim

.PHONY: mod
mod:
	go mod tidy

.PHONY: install
install:
	go install github.com/mgechev/revive@latest

.PHONY: revive
revive:
	revive -set_exit_status ./...

.PHONY: build
lint: revive
	golangci-lint run

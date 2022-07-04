all: build

build: linux mac

linux:
	GOOS=linux GOARCH=amd64 go build -o bin/qconf cmd/main.go

mac:
	GOOS=darwin GOARCH=amd64 go build -o bin/qconf_darwin cmd/main.go

lint:
	staticcheck ./...


BINARY = imServerGo
GOARCH = amd64
RELEASE?=0.0.0.1

COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Release=${RELEASE} -X main.Commit=${COMMIT} -X main.BuildTime=${BUILD_TIME}"

.PHONY: default
default:
	cp cmd/server/app-example.toml release/server/app.toml
	go build ${LDFLAGS} -o release/server/${BINARY} cmd/server/main.go

.PHONY: client
client:
	cp cmd/client/app-example.toml release/client/app.toml
	go build ${LDFLAGS} -o release/client/${BINARY} cmd/client/main.go

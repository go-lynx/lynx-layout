GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
GOBIN:=$(shell go env GOBIN)
VERSION=$(shell git describe --tags --always)

# Ensure installed tools are in PATH for current shell
export PATH := $(PATH):$(GOPATH)/bin:$(GOBIN)

ifeq ($(GOHOSTOS), windows)
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
	# Discovery for wire.go directories
	WIRE_DIRS=$(shell $(Git_Bash) -c "find cmd -name wire.go -exec dirname {} \;" | sort -u)
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
	# Discovery for wire.go directories
	WIRE_DIRS=$(shell find cmd -name wire.go -exec dirname {} \; | sort -u)
endif

.PHONY: init
# init env
init:
	go install github.com/go-lynx/lynx/cmd/lynx@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto (OpenAPI output to docs/openapi.yaml for use with lynx-swagger)
api:
	mkdir -p docs
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:./docs \
 	       $(API_PROTO_FILES)

.PHONY: validate
# generate validate proto
validate:
	protoc --proto_path=. \
		   --proto_path=./third_party \
		   --go_out=paths=source_relative:. \
		   --validate_out=paths=source_relative,lang=go:. \
		   $(API_PROTO_FILES)

.PHONY: build
# build (Build binary, depends on wire generation)
build: wire
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: test
# test
test:
	go test ./...

.PHONY: generate
# generate
generate:
	go mod tidy
	go generate ./...

.PHONY: wire
# generate wire
wire:
	@for dir in $(WIRE_DIRS); do \
		echo "Running wire in $$dir..."; \
		cd $$dir && go run -mod=mod github.com/google/wire/cmd/wire; \
	done

.PHONY: all
# all (One-stop generation and build)
all:
	$(MAKE) api
	$(MAKE) config
	$(MAKE) generate
	$(MAKE) ent
	$(MAKE) validate
	$(MAKE) build

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

.PHONY: ent
# generate ent
ent:
	go generate ./internal/data/ent

# build and run
run:
	$(MAKE) all
	kratos run

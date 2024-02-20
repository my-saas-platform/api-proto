GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

MAKE_FILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CURRENT_ABS_PATH=$(shell dirname $(MAKE_FILE_PATH))
CURRENT_PATH=$(shell basename $(CURRENT_ABS_PATH))
PROJECT_PATH=$(shell echo "../../")
APP_RELATIVE_PATH=$(shell a=`basename $$PWD` && echo $${a})
ifeq ($(APP_RELATIVE_PATH), $(CURRENT_PATH))
	PROJECT_PATH=./
	APP_RELATIVE_PATH=
endif

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	GIT_BASH= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	COMMON_PROTO_FILES=$(shell $(GIT_BASH) -c "find $(PROJECT_PATH)api/common -name *.proto")
else
endif

.DEFAULT_GOAL := help

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.PHONY: init
# init and install necessary software
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest

# include
include api/makefile_protoc.mk
# servers
include api/config/makefile_protoc.mk
include api/ping-service/makefile_protoc.mk

.PHONY: echo
# echo test content
echo:
	@echo $(CURRENT_PATH)

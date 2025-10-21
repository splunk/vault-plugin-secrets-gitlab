NAME?=vault-plugin-secrets-gitlab

.DEFAULT_GOAL := all
all: get build lint test 

get:
	go get ./...

build:
	go build -v -o plugins/$(NAME)

build-linux:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o plugins/$(NAME)


lint: .tools/golangci-lint
	.tools/golangci-lint --version
	.tools/golangci-lint run -v

test:
	mkdir -p coverage/unit
	go test -short -parallel=10 -v -covermode=count -cover ./... $(TESTARGS) -args -test.gocoverdir="$(shell pwd)/coverage/unit"

# acc-test: tools
#   mkdir -p coverage/int
# 	@(eval $$(./scripts/init_dev.sh) && go test -parallel=10 -v -covermode=count -cover ./... -run=TestAcc -args -test.gocoverdir="$(shell pwd)/coverage/int")

report: .tools/gocover-cobertura
	mkdir -p coverage
	# go tool covdata textfmt -i ./coverage/unit,./coverage/int -o coverage/profile
	go tool covdata textfmt -i ./coverage/unit -o coverage/profile
	go tool cover -func=coverage/profile
	go tool cover -html=coverage/profile -o coverage/coverage.html
	.tools/gocover-cobertura < coverage/profile > coverage/coverage.xml

vault-only: build
	vault server -log-level=debug -dev -dev-root-token-id=root -dev-plugin-dir=./plugins

dev: tools build-linux
	@./scripts/init_dev.sh

clean-dev:
	@cd scripts && docker-compose down

clean-all: clean-dev
	@rm -rf .tools coverage*.* plugins

tools: .tools .tools/docker-compose .tools/gocover-cobertura .tools/golangci-lint .tools/jq .tools/vault

.tools:
	@mkdir -p .tools

.tools/docker-compose: DOCKER_COMPOSE_VERSION = 2.40.0
.tools/docker-compose: DOCKER_COMPOSE_BINARY = "docker-compose-$(shell uname -s)-$(shell uname -m)"
.tools/docker-compose:
	curl -so .tools/docker-compose -L "https://github.com/docker/compose/releases/download/$(DOCKER_COMPOSE_VERSION)/$(DOCKER_COMPOSE_BINARY)"
	@chmod +x .tools/docker-compose

.tools/gocover-cobertura:
	export GOBIN=$(shell pwd)/.tools; go install github.com/boumenot/gocover-cobertura@v1.4.0

.tools/golangci-lint:
	export GOBIN=$(shell pwd)/.tools; go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0

.tools/jq: JQ_VERSION = 1.8.1
.tools/jq: JQ_PLATFORM = $(patsubst darwin,osx-amd,$(shell uname -s | tr A-Z a-z))
.tools/jq:
	curl -so .tools/jq -sSL https://github.com/stedolan/jq/releases/download/jq-$(JQ_VERSION)/jq-$(JQ_PLATFORM)64
	@chmod +x .tools/jq

.tools/vault: VAULT_VERSION = 1.19.5
.tools/vault: VAULT_PLATFORM = $(shell uname -s | tr A-Z a-z)
.tools/vault:
	curl -so .tools/vault.zip -sSL https://releases.hashicorp.com/vault/$(VAULT_VERSION)/vault_$(VAULT_VERSION)_$(VAULT_PLATFORM)_amd64.zip
	(cd .tools && unzip -o vault.zip && rm vault.zip)

.PHONY: all get build build-linux publish lint test test-artacc test-vaultacc report vault-only dev clean-dev clean-all tools

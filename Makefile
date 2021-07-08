NAME?=vault-gitlab-token-plugin

.DEFAULT_GOAL := all
all: get build lint test 

get:
	go get ./...

build:
	go build -v -o plugins/$(NAME)

build-linux:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o plugins/$(NAME)

# publish:
# 	$(eval VERSION=$(shell gitversion show))
# 	./scripts/publish.sh linux amd64 $(VERSION)

lint: .tools/golangci-lint
	.tools/golangci-lint run

test:
	go test -short -parallel=10 -v -covermode=count -coverprofile=coverage_unit.out ./... $(TESTARGS)

# acceptance-test: tools
# 	@(eval $$(./scripts/init_dev.sh) && go test -parallel=10 -v -covermode=count -coverprofile=coverage_artacc.out ./... -run=TestAcc)

report: .tools/gocover-cobertura .tools/gocovmerge
	.tools/gocovmerge coverage_*.out > coverage.out
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	.tools/gocover-cobertura < coverage.out > coverage.xml

vault-only: build
	vault server -log-level=debug -dev -dev-root-token-id=root -dev-plugin-dir=./plugins

dev: tools build-linux
	@./scripts/init_dev.sh

clean-dev:
	@cd scripts && docker-compose down

clean-all: clean-dev
	@rm -rf .tools coverage*.* plugins

tools: .tools .tools/docker-compose .tools/gocover-cobertura .tools/gocovmerge .tools/golangci-lint .tools/jq .tools/vault

.tools:
	@mkdir -p .tools

.tools/docker-compose: DOCKER_COMPOSE_VERSION = 1.29.1
.tools/docker-compose: DOCKER_COMPOSE_BINARY = "docker-compose-$(shell uname -s)-$(shell uname -m)"
.tools/docker-compose:
	curl -so .tools/docker-compose -L "https://github.com/docker/compose/releases/download/$(DOCKER_COMPOSE_VERSION)/$(DOCKER_COMPOSE_BINARY)"
	@chmod +x .tools/docker-compose

.tools/gocover-cobertura:
	export GOBIN=$(shell pwd)/.tools; go install github.com/boumenot/gocover-cobertura@v1.1.0

.tools/gocovmerge:
	export GOBIN=$(shell pwd)/.tools; go install github.com/wadey/gocovmerge@master

.tools/golangci-lint:
	export GOBIN=$(shell pwd)/.tools; go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0

.tools/jq: JQ_VERSION = 1.6
.tools/jq: JQ_PLATFORM = $(patsubst darwin,osx-amd,$(shell uname -s | tr A-Z a-z))
.tools/jq:
	curl -so .tools/jq -sSL https://github.com/stedolan/jq/releases/download/jq-$(JQ_VERSION)/jq-$(JQ_PLATFORM)64
	@chmod +x .tools/jq

.tools/vault: VAULT_VERSION = 1.7.1
.tools/vault: VAULT_PLATFORM = $(shell uname -s | tr A-Z a-z)
.tools/vault:
	curl -so .tools/vault.zip -sSL https://releases.hashicorp.com/vault/$(VAULT_VERSION)/vault_$(VAULT_VERSION)_$(VAULT_PLATFORM)_amd64.zip
	(cd .tools && unzip -o vault.zip && rm vault.zip)

.PHONY: all get build build-linux publish lint test test-artacc test-vaultacc report vault-only dev clean-dev clean-all tools

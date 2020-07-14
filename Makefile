.DEFAULT_GOAL	:= build

#------------------------------------------------------------------------------
# Variables
#------------------------------------------------------------------------------

SHELL		:= /bin/bash
PKG			:= github.com/disneystreaming/ssm-helpers
IMAGE		:= docker.pkg.github.com/disneystreaming/ssm-helpers/ssm

# Pure Go sources (not vendored and not generated)
GOFILES		= $(shell find . -type f -name '*.go' )
GODIRS		= $(shell go list -f '{{.Dir}}' ./...)

# Current platforms os and architecture
GOOS		= $(shell go env GOOS)
GOARCH		= $(shell go env GOARCH)

.PHONY: build
build: format
	@echo "--> building"
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=0 go build -o bin/ssm main.go

.PHONY: docker
docker: format
	@echo "--> building linux binary"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/ssm main.go
	@echo "--> building docker image"
	@docker build . -f Dockerfile -t $(IMAGE)

.PHONY: clean
clean:
	@echo "--> cleaning compiled objects and binaries"
	@go clean -tags netgo -i ./...

.PHONY: test
test:
	@echo "--> running unit tests"
	@go test -v $(GODIRS)

.PHONY: check
check: format.check vet lint

.PHONY: cover
cover:
	@echo "--> generating coverage report"
	@go test -v $(GODIRS) -cover

.PHONY: format
format: tools.goimports
	@echo "--> formatting code with 'goimports' tool"
	@goimports -local $(PKG) -w -l $(GOFILES)

.PHONY: format.check
format.check: tools.goimports
	@echo "--> checking code formatting with 'goimports' tool"
	@goimports -local $(PKG) -l $(GOFILES) | sed -e "s/^/\?\t/" | tee >(test -z)

.PHONY: vet
vet: tools.govet
	@echo "--> checking code correctness with 'go vet' tool"
	@go vet ./...

.PHONY: lint
lint: tools.golint
	@echo "--> checking code style with 'golint' tool"
	@echo $(GODIRS) | xargs -n 1 golint

#---------------
#-- tools
#---------------
.PHONY: tools tools.goimports tools.golint tools.govet

tools: tools.goimports tools.golint tools.govet

tools.goimports:
	@command -v goimports >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "--> installing goimports"; \
		go get golang.org/x/tools/cmd/goimports; \
	fi

tools.govet:
	@go tool vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		echo "--> installing govet"; \
		go get golang.org/x/tools/cmd/vet; \
	fi

tools.golint:
	@command -v golint >/dev/null ; if [ $$? -ne 0 ]; then \
		echo "--> installing golint"; \
		go get -u golang.org/x/lint/golint; \
	fi

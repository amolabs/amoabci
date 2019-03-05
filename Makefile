.PHONY: build docker run-cluster test

all: install

GO := $(shell command -v go 2> /dev/null)
FS := /
# go source code files including files from vendor directory
GOSRCS=$(shell find . -name \*.go)
#BUILDENV=CGO_ENABLED=0

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

ifneq ($(TARGET),)
  BUILDENV += GOOS=$(TARGET)
endif

GOPATH ?= $(shell $(GO) env GOPATH)
GITHUBDIR := $(GOPATH)$(FS)src$(FS)github.com

GOPATH ?= $(shell $(GO) env GOPATH)

go_get = $(if $(findstring Windows_NT,$(OS)),\
IF NOT EXIST $(GITHUBDIR)$(FS)$(1)$(FS) ( mkdir $(GITHUBDIR)$(FS)$(1) ) else (cd .) &\
IF NOT EXIST $(GITHUBDIR)$(FS)$(1)$(FS)$(2)$(FS) ( cd $(GITHUBDIR)$(FS)$(1) && git clone https://github.com/$(1)/$(2) ) else (cd .) &\
,\
mkdir -p $(GITHUBDIR)$(FS)$(1) &&\
(test ! -d $(GITHUBDIR)$(FS)$(1)$(FS)$(2) && cd $(GITHUBDIR)$(FS)$(1) && git clone https://github.com/$(1)/$(2)) || true &&\
)\
cd $(GITHUBDIR)$(FS)$(1)$(FS)$(2) && git fetch origin && git checkout -q $(3)

go_install = $(call go_get,$(1),$(2),$(3)) && cd $(GITHUBDIR)$(FS)$(1)$(FS)$(2) && $(GO) install

tags: $(GOSRCS)
	gotags -R -f tags .

#tools: $(GOPATH)/bin/dep $(GOPATH)/bin/gometalinter $(GOPATH)/bin/statik $(GOPATH)/bin/goimports
get_tools: $(GOPATH)/bin/dep

$(GOPATH)/bin/dep:
	$(call go_get,golang,dep,22125cfaa6ddc71e145b1535d4b7ee9744fefff2)
	cd $(GITHUBDIR)$(FS)golang$(FS)dep$(FS)cmd$(FS)dep && $(GO) install

#v2.0.11
$(GOPATH)/bin/gometalinter:
	$(call go_install,alecthomas,gometalinter,17a7ffa42374937bfecabfb8d2efbd4db0c26741)

$(GOPATH)/bin/statik:
	$(call go_install,rakyll,statik,v0.1.5)

$(GOPATH)/bin/goimports:
	go get golang.org/x/tools/cmd/goimports

get_vendor_deps:
	@echo "--> Generating vendor directory via dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -vendor-only

update_vendor_deps:
	@echo "--> Running dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -update

build:
	@echo "--> Building amo daemon (amod)"
	$(BUILDENV) go build ./cmd/amod
	@echo "--> Building amo console (amocli)"
	$(BUILDENV) go build ./cmd/amocli

install:
	@echo "--> Installing amo daemon (amod)"
	go install ./cmd/amod
	@echo "--> Installing amo console (amocli)"
	go install ./cmd/amocli

test:
	go test ./...

docker:
	$(MAKE) TARGET=linux build
	docker build -t amolabs/amod .

run-cluster: docker
	docker-compose -p $(USER) up -d

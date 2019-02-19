.PHONY: build docker run-cluster

all: build

GO := $(shell command -v go 2> /dev/null)
FS := /
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

#tools: $(GOPATH)/bin/dep $(GOPATH)/bin/gometalinter $(GOPATH)/bin/statik $(GOPATH)/bin/goimports
tools: $(GOPATH)/bin/dep

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

vendor-deps:
	@echo "--> Generating vendor directory via dep ensure"
	@rm -rf .vendor-new
	@dep ensure -v -vendor-only

build:
	@echo "--> Building amo daemon"
	$(BUILDENV) go build -a -o amod .

docker:
	docker build -t amod .

run-cluster: docker
	docker-compose up

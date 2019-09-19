.PHONY: build docker run-cluster test

all: build

GO := $(shell command -v go 2> /dev/null)
FS := /
# go source code files
GOSRCS=$(shell find . -name \*.go)
BUILDENV=CGO_ENABLED=0

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

ifneq ($(TARGET),)
  BUILDENV += GOOS=$(TARGET)
endif

GOPATH ?= $(shell $(GO) env GOPATH)
GITHUBDIR := $(GOPATH)$(FS)src$(FS)github.com
TMPATH=$(GOPATH)/src/github.com/tendermint/tendermint

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

#v2.0.11
$(GOPATH)/bin/gometalinter:
	$(call go_install,alecthomas,gometalinter,17a7ffa42374937bfecabfb8d2efbd4db0c26741)

$(GOPATH)/bin/statik:
	$(call go_install,rakyll,statik,v0.1.5)

$(GOPATH)/bin/goimports:
	go get golang.org/x/tools/cmd/goimports

build:
	@echo "--> Building amo daemon (amod)"
	$(BUILDENV) go build ./cmd/amod

install:
	@echo "--> Installing amo daemon (amod)"
	$(BUILDENV) go install ./cmd/amod

test:
	go test ./...

tendermint:
	-git clone https://github.com/tendermint/tendermint $(TMPATH)
	cd $(TMPATH); git checkout v0.32.3
	make -C $(TMPATH) tools
	make -C $(TMPATH) build-linux
	cp $(TMPATH)/build/tendermint ./

docker: tendermint
	$(MAKE) TARGET=linux build
	cp -f amod tendermint DOCKER/
	docker build -t amolabs/amod DOCKER

clean:
	rm -f amod tendermint

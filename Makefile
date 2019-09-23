.PHONY: build docker run-cluster test

all: build

GO := $(shell command -v go 2> /dev/null)
# go source code files
GOSRCS=$(shell find . -name \*.go)
BUILDENV=CGO_ENABLED=1

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

ifneq ($(TARGET),)
  BUILDENV += GOOS=$(TARGET)
endif

tags: $(GOSRCS)
	gotags -R -f tags .

build:
	@echo "--> Building amo daemon (amod)"
	$(BUILDENV) go build -tags "cleveldb" ./cmd/amod

install:
	@echo "--> Installing amo daemon (amod)"
	$(BUILDENV) go install -tags "cleveldb" ./cmd/amod

test:
	go test ./...

docker:
	docker build -t amolabs/amod DOCKER

clean:
	rm -f amod

.PHONY: build docker run-cluster test

all: build

GO := $(shell command -v go 2> /dev/null)
# go source code files
GOSRCS=$(shell find . -name \*.go)
BUILDENV=
BUILDTAGS=

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

ifneq ($(TARGET),)
  BUILDENV += GOOS=$(TARGET)
endif

PROFCMD=go test -cpuprofile=cpu.prof -bench .

tags: $(GOSRCS)
	gotags -R -f tags .

build:
	@echo "--> Building amo daemon (amod)"
	$(BUILDENV) go build ./cmd/amod
	$(BUILDENV) go build ./cmd/repair

build_c: l_BUILDENV = $(BUILDENV) CGO_ENABLED=1
build_c:
	@echo "--> Building amo daemon (amod)"
	$(l_BUILDENV) go build -tags cleveldb,rocksdb ./cmd/amod
	$(l_BUILDENV) go build -tags cleveldb,rocksdb ./cmd/repair

install:
	@echo "--> Installing amo daemon (amod)"
	$(BUILDENV) go install ./cmd/amod

install_c: l_BUILDENV = $(BUILDENV) CGO_ENABLED=1
install_c:
	@echo "--> Installing amo daemon (amod)"
	$(l_BUILDENV) go install -tags cleveldb,rocksdb ./cmd/amod

test:
	go test ./...

test_c: l_BUILDENV = $(BUILDENV) CGO_ENABLED=1
test_c:
	$(l_BUILDENV) go test -tags cleveldb,rocksdb ./...

bench:
	cd amo; $(PROFCMD)
	cd amo/store; $(PROFCMD)

docker:
	COPYFILE_DISABLE=true tar zcf amoabci-docker.tar.gz Makefile go.mod go.sum cmd amo crypto Dockerfile DOCKER
	docker build -t amolabs/amod - < amoabci-docker.tar.gz

clean:
	rm -f amod *.test

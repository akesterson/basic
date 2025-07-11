VERSION:=0.2
SRCFILES:=$(shell find . -type f -maxdepth 1 -name '*.go')
OS:=$(shell uname -o)
ARCH:=$(shell uname -m)

ifeq ($(ARCH),x86_64)
	GO_ARCH=amd64
else
	GO_ARCH=$(ARCH)
endif

ifeq ($(OS),Msys)
	EXE_EXT:=.exe
	GO_OS=windows
	BUILD=CGO_ENABLED=1 CC=gcc GOOS=$(GO_OS) GOARCH=$(GO_ARCH) "$(GO)" build -o basic$(EXE_EXT) $(SRCFILES)
else
	EXE_EXT:=
ifeq ($(OS),darwin)
	GO_OS=darwin
else
	GO_OS:=linux
endif
	BUILD=CGO_ENABLED=1 CC=gcc GOOS=$(GO_OS) GOARCH=$(GO_ARCH) "$(GO)" build -tags static -ldflags "-s -w" -o basic$(EXE_EXT) $(SRCFILES)
endif

DISTFILE:=basic$(EXE_EXT)
GO:=$(shell which go$(EXE_EXT))

.PHONY: clean
.PHONY: tests

all: $(DISTFILE)

clean:
	rm -fr $(DISTFILE) release/

tests:
	bash ./test.sh

$(DISTFILE): $(SRCFILES)
	$(BUILD)

release: release/$(GO_OS)/$(DISTFILE)

release/windows/$(DISTFILE): $(DISTFILE)
	mkdir -p release/windows
	cp $$(ldd $(DISTFILE) | cut -d '>' -f 2 | cut -d '(' -f 1 | grep -vi /windows/system) release/windows/
	cp $(DISTFILE) release/windows/$(DISTFILE)
	cd release/windows && zip basic-$(GO_OS)-$(GO_ARCH)-$(VERSION).zip basic.exe *dll

release/linux/$(DISTFILE): $(DISTFILE)
	mkdir -p release/linux
	cp $(DISTFILE) release/linux/$(DISTFILE)
	cd release/linux
	tar -czvf $(DISTFILE)-$(GO_OS)-$(GO_ARCH)-$(VERSION).tar.gz $(DISTFILE)

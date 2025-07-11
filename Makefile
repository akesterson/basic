SRCFILES:=$(shell find . -type f -maxdepth 1 -name '*.go')
DISTFILE:=basic.exe
OS:=$(shell uname -o)

ifeq ($(OS),Msys)
	EXE_EXT:=.exe
	GO_OS=windows
else
	EXE_EXT:=
	GO_OS=linux
endif

GO:=$(shell which go$(EXE_EXT))

.PHONY: clean
.PHONY: tests

all: $(DISTFILE)

clean:
	rm -fr $(DISTFILE)

tests:
	bash ./test.sh

$(DISTFILE): $(SRCFILES)
	CGO_ENABLED=1 CC=gcc GOOS=$(GO_OS) GOARCH=amd64 $(GO) build -tags static -ldflags "-s -w" -o basic$(EXE_EXT) $(SRCFILES)

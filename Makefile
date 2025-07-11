SRCFILES:=$(shell find . -type f -maxdepth 1 -name '*.go')
DISTFILE:=basic.exe
OS:=$(shell uname -o)

ifeq ($(OS),Msys)
	EXE_EXT:=.exe
	GO_OS=windows
	BUILD=CGO_ENABLED=1 CC=gcc GOOS=$(GO_OS) GOARCH=amd64 "$(GO)" build -o basic$(EXE_EXT) $(SRCFILES)

else
	EXE_EXT:=
	GO_OS=linux
	BUILD=CGO_ENABLED=1 CC=gcc GOOS=$(GO_OS) GOARCH=amd64 "$(GO)" build -tags static -ldflags "-s -w" -o basic$(EXE_EXT) $(SRCFILES)
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
	$(BUILD)

release: release/$(GO_OS)/$(DISTFILE)

release/windows/$(DISTFILE): $(DISTFILE)
	mkdir -p release/windows
	cp $$(ldd $(DISTFILE) | cut -d '>' -f 2 | cut -d '(' -f 1 | grep -vi /windows/system) release/windows/
	cp $(DISTFILE) release/windows/$(DISTFILE)

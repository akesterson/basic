SRCFILES:=$(shell find . -type f -maxdepth 1 -name '*.go')
DISTFILE:=basic.exe
OS:=$(shell uname -o)

ifeq ($(OS),Msys)
	EXE_EXT:=.exe
else
	EXE_EXT:=
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
	$(GO) build -o basic$(EXE_EXT) $(SRCFILES)

SRCFILES:=$(shell find . -type f -maxdepth 1 -name '*.go')
DISTFILE:=basic.exe
OS:=$(shell uname -o)
SDL2_INCLUDE:="-I/mingw64/include"
SDL2_LIB:="-L/mingw64/lib -lSDL2"

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
	CGO_CFLAGS=$(SDL2_INCLUDE) CGO_LDFLAGS=$(SDL2_LIB) $(GO) build -o basic$(EXE_EXT) $(SRCFILES)

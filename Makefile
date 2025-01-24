SRCFILES:=$(shell find . -type f -maxdepth 1 -name '*.go')
DISTFILE:=basic.exe

.PHONY: clean

all: $(DISTFILE)

clean:
	rm -fr $(DISTFILE)

$(DISTFILE): $(SRCFILES)
	go build -o basic.exe $(SRCFILES)

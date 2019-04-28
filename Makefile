export GO111MODULE=on

SOURCES = main.go
OUTPUT = bin

GO = go
GOGET = go get -u

all: $(OUTPUT)/goxel

$(OUTPUT)/goxel: $(SOURCES)
	$(GO) build -x -o $(OUTPUT)/goxel $(SOURCES)

deps:
	${GO} build -v -o ${OUTPUT}/goxel

clean:
	$(GO) clean -x
	rm -rf $(OUTPUT)/goxel

test:
	cd goxel && $(GO) test

install:
	$(GO) install -v .

.PHONY: all clean install deps

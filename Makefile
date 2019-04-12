SOURCES = main.go
OUTPUT = bin

GO = go
GOGET = go get -u

all: deps $(OUTPUT)/goxel

$(OUTPUT)/goxel: $(SOURCES)
	$(GO) build -x -o $(OUTPUT)/goxel $(SOURCES)

deps:
	$(GOGET) github.com/dustin/go-humanize

clean:
	$(GO) clean -x
	rm -rf $(OUTPUT)/goxel

install:
	$(GO) install -v .

.PHONY: all clean install deps

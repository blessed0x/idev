BIN     := idev

GOFMT_SRC := cmd device internal

.PHONY: build test vet tool-check gofmt-check lint qa install clean

build:
	go build -o bin/$(BIN) ./cmd/idev

test:
	go test ./...

vet:
	go vet ./...

tool-check:
	go mod tidy -diff

gofmt-check:
	@out=$$(gofmt -l $(GOFMT_SRC)); \
	if [ -n "$$out" ]; then \
	  echo "gofmt needed on:"; echo "$$out"; exit 1; \
	fi

lint: gofmt-check vet tool-check
	go run honnef.co/go/tools/cmd/staticcheck ./...
	go run golang.org/x/vuln/cmd/govulncheck ./...

qa: lint
	go test -race ./...
	@echo "=== cross-compile matrix ==="
	@for os in darwin linux windows; do \
	  for arch in arm64 amd64; do \
	    echo "  GOOS=$$os GOARCH=$$arch"; \
	    GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build ./... || exit 1; \
	  done; \
	done

install:
	go install ./cmd/idev

clean:
	rm -rf bin dist

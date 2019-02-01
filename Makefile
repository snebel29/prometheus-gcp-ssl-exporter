BINARY_VERSION=0.0.1
LD_FLAGS="-X github.com/snebel29/prometheus-gcp-ssl-exporter/internal/pkg/cli.Version=$(BINARY_VERSION)"

build: deps
	go build -ldflags $(LD_FLAGS) cmd/*.go
test:
	go test -v ./...
clean:
	go clean
deps:
	dep ensure -v

.PHONY: build test clean

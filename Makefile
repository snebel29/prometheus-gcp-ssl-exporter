BINARY_VERSION=0.0.1
BINARY_NAME=prometheus-gcp-ssl-exporter
LD_FLAGS="-X github.com/snebel29/prometheus-gcp-ssl-exporter/internal/pkg/cli.Version=$(BINARY_VERSION)"

build: test
	go build -ldflags $(LD_FLAGS) -o $(BINARY_NAME) cmd/*.go
test:
	go test -v ./...
clean:
	go clean
deps:
	dep ensure -v
run:
	./$(BINARY_NAME)

.PHONY: build test clean run

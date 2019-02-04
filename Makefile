GOOGLE_APPLICATION_CREDENTIALS=$(realpath internal/pkg/collector/govcr-fixtures/application_default_credentials.json)
BINARY_VERSION=0.0.1
LD_FLAGS="-X github.com/snebel29/prometheus-gcp-ssl-exporter/internal/pkg/cli.Version=$(BINARY_VERSION)"

build: deps
	go build -ldflags $(LD_FLAGS) cmd/*.go
test:
	$(info GOOGLE_APPLICATION_CREDENTIALS=$(GOOGLE_APPLICATION_CREDENTIALS))
	go test -v ./...
clean:
	go clean
deps:
	dep ensure -v

.PHONY: build test clean

VERSION=x.x.x-development
REPOSITORY=github.com/snebel29/prometheus-gcp-ssl-exporter
GOOGLE_APPLICATION_CREDENTIALS=$(realpath internal/pkg/collector/govcr-fixtures/application_default_credentials.json)
LD_FLAGS="-X ${REPOSITORY}/internal/pkg/cli.Version=$(VERSION)"

export GOOGLE_APPLICATION_CREDENTIALS

build: deps
	go build -ldflags $(LD_FLAGS) cmd/*.go
test:
	$(info GOOGLE_APPLICATION_CREDENTIALS=$(GOOGLE_APPLICATION_CREDENTIALS))
	go test -v ./...
clean:
	go clean
deps:
	dep ensure -v
container:
	docker build -f build/Dockerfile \
		--build-arg VERSION=$(VERSION) \
		--build-arg REPOSITORY=$(REPOSITORY) \
		-t prometheus-gcp-ssl-exporter:latest \
		-t prometheus-gcp-ssl-exporter:$(VERSION) .

.PHONY: build test clean container

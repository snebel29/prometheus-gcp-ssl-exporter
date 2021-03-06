VERSION=x.x.x-development
REPOSITORY=github.com/snebel29/prometheus-gcp-ssl-exporter
GOOGLE_APPLICATION_CREDENTIALS=$(realpath internal/pkg/collector/govcr-fixtures/application_default_credentials.json)
LD_FLAGS="-X ${REPOSITORY}/internal/pkg/cli.Version=$(VERSION) -w -extldflags -static"

export GOOGLE_APPLICATION_CREDENTIALS

build: deps
	CGO_ENABLED=0 go build -ldflags $(LD_FLAGS) cmd/*.go
test:
	go test -v ./...
clean:
	go clean
deps:
	dep ensure -v

docker-image:
	docker build -f build/Dockerfile \
		--build-arg VERSION=$(VERSION) \
		--build-arg REPOSITORY=$(REPOSITORY) \
		-t snebel29/prometheus-gcp-ssl-exporter:latest \
		-t snebel29/prometheus-gcp-ssl-exporter:$(VERSION) .

publish-docker-image:
	docker push snebel29/prometheus-gcp-ssl-exporter:$(VERSION)
	docker push snebel29/prometheus-gcp-ssl-exporter:latest

.PHONY: build test clean docker-image publish-docker-image

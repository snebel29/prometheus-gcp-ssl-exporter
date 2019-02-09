

FROM golang:1.11.5-alpine3.8 AS build

ARG REPOSITORY
ARG VERSION

RUN apk --update add --no-cache git make ca-certificates
RUN go get github.com/golang/dep/cmd/dep

COPY . /go/src/${REPOSITORY}
WORKDIR /go/src/${REPOSITORY}

RUN echo ${REPOSITORY}; make VERSION=${VERSION} build

FROM scratch
ARG REPOSITORY
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/${REPOSITORY}/prometheus-gcp-ssl-exporter /prometheus-gcp-ssl-exporter
ENTRYPOINT ["/prometheus-gcp-ssl-exporter"]



FROM golang:1.11.5-alpine3.8 AS build

ARG REPOSITORY
ARG VERSION

RUN apk add --no-cache git make
RUN go get github.com/golang/dep/cmd/dep

COPY . /go/src/${REPOSITORY}
WORKDIR /go/src/${REPOSITORY}
RUN dep ensure -vendor-only

RUN make VERSION= build

FROM scratch
ARG REPOSITORY
COPY --from=build /go/src/${REPOSITORY}/prometheus-gcp-ssl-exporter /
ENTRYPOINT ["/prometheus-gcp-ssl-exporter"]

# Prometheus-gcp-ssl-exporter
Export your attributes of your TLS/SSL certificates within Google Cloud Platform, currently only the `NotAfter` field of every certificate transformed to seconds left to expire, example below

```
# HELP gcp_ssl_validity_seconds Time for an ssl certificate to expire
# TYPE gcp_ssl_validity_seconds gauge
gcp_ssl_validity_seconds{name="star-mycertificate",project="my-gcpp-project"} 5.8653036e+07
...
```

## What is this for?
You can monitor all your GCP hosted certificate expiration time in an straighforward way, without the need to setup external probes or having any prior information about them.

## What this is not for?
A replacement for external blackbox monitoring on your urls, also this won't tell you if those certificates are in use and this won't monitor applications doing their own TLS termination.

## Install
```
$ go get -u github.com/snebel29/prometheus-gcp-ssl-exporter/cmd
```

## Usage
```
usage: prometheus-gcp-ssl-exporter --project=PROJECT [<flags>]

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
  -m, --metrics-path="/metrics"  URI path where metrics will be exposed
      --port="8888"              Port to listen on
  -p, --project=PROJECT ...      GCP project where to fetch certificates from
      --version                  Show application version.
```

### Example
```
$ prometheus-gcp-ssl-exporter -p my-project1 -p my-project2
```

## Development

### Build
```
$ make build
```

### Test
```
$ make test
```

### Synchronize vendor
```
$ make deps
```


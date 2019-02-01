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

## Authentication
The exporter needs to authenticate and be authorized to do `compute.sslCertificates.get` and `compute.sslCertificates.list` within the Google cloud compute API, to do so Google offer several [methods to authenticate for production workloads](https://cloud.google.com/docs/authentication/production) from which creating a service account is common, in a nutshell you could create a service account with the least privilege principle like this:

Create custom role
```
$ gcloud iam roles create sslViewer \
	--project ${PROJECT_ID} \
	--title "Compute SSL Viewer" \
	--description "List and Get SSL certificates" \
	--stage GA \
	--permissions compute.sslCertificates.get,compute.sslCertificates.list
```

Create service account
```
$ gcloud iam service-accounts create ${NAME}
$ gcloud projects add-iam-policy-binding ${PROJECT_ID} --member "serviceAccount:${NAME}@${PROJECT_ID}.iam.gserviceaccount.com" --role "projects/${PROJECT_ID}/roles/sslViewer"
$ gcloud iam service-accounts keys create ${FILE_NAME}.json --iam-account ${NAME}@${PROJECT_ID}.iam.gserviceaccount.com
```

Then create `GOOGLE_APPLICATION_CREDENTIALS` environment variable pointing to the credentials file.
```
$ export GOOGLE_APPLICATION_CREDENTIALS="/home/user/Downloads/${FILE_NAME}.json"
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

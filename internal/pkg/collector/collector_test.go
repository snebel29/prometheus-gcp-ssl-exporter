package collector

import (
	"fmt"
	"errors"
	"testing"
	"github.com/seborama/govcr"
	"github.com/prometheus/client_golang/prometheus"
)

var pemData = `-----BEGIN CERTIFICATE-----
MIIDujCCAqKgAwIBAgIIE31FZVaPXTUwDQYJKoZIhvcNAQEFBQAwSTELMAkGA1UE
BhMCVVMxEzARBgNVBAoTCkdvb2dsZSBJbmMxJTAjBgNVBAMTHEdvb2dsZSBJbnRl
cm5ldCBBdXRob3JpdHkgRzIwHhcNMTQwMTI5MTMyNzQzWhcNMTQwNTI5MDAwMDAw
WjBpMQswCQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTEWMBQGA1UEBwwN
TW91bnRhaW4gVmlldzETMBEGA1UECgwKR29vZ2xlIEluYzEYMBYGA1UEAwwPbWFp
bC5nb29nbGUuY29tMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfRrObuSW5T7q
5CnSEqefEmtH4CCv6+5EckuriNr1CjfVvqzwfAhopXkLrq45EQm8vkmf7W96XJhC
7ZM0dYi1/qOCAU8wggFLMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAa
BgNVHREEEzARgg9tYWlsLmdvb2dsZS5jb20wCwYDVR0PBAQDAgeAMGgGCCsGAQUF
BwEBBFwwWjArBggrBgEFBQcwAoYfaHR0cDovL3BraS5nb29nbGUuY29tL0dJQUcy
LmNydDArBggrBgEFBQcwAYYfaHR0cDovL2NsaWVudHMxLmdvb2dsZS5jb20vb2Nz
cDAdBgNVHQ4EFgQUiJxtimAuTfwb+aUtBn5UYKreKvMwDAYDVR0TAQH/BAIwADAf
BgNVHSMEGDAWgBRK3QYWG7z2aLV29YG2u2IaulqBLzAXBgNVHSAEEDAOMAwGCisG
AQQB1nkCBQEwMAYDVR0fBCkwJzAloCOgIYYfaHR0cDovL3BraS5nb29nbGUuY29t
L0dJQUcyLmNybDANBgkqhkiG9w0BAQUFAAOCAQEAH6RYHxHdcGpMpFE3oxDoFnP+
gtuBCHan2yE2GRbJ2Cw8Lw0MmuKqHlf9RSeYfd3BXeKkj1qO6TVKwCh+0HdZk283
TZZyzmEOyclm3UGFYe82P/iDFt+CeQ3NpmBg+GoaVCuWAARJN/KfglbLyyYygcQq
0SgeDh8dRKUiaW3HQSoYvTvdTuqzwK4CXsr3b5/dAOY8uMuG/IAR3FgwTbZ1dtoW
RvOTa8hYiU6A475WuZKyEHcwnGYe57u2I2KbMgcKjPniocj4QzgYsVAVKW3IwaOh
yE+vPxsiUkvQHdO2fojCkY8jg70jxM+gu59tPDNbw3Uh/2Ij310FgTHsnGQMyA==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEBDCCAuygAwIBAgIDAjppMA0GCSqGSIb3DQEBBQUAMEIxCzAJBgNVBAYTAlVT
MRYwFAYDVQQKEw1HZW9UcnVzdCBJbmMuMRswGQYDVQQDExJHZW9UcnVzdCBHbG9i
YWwgQ0EwHhcNMTMwNDA1MTUxNTU1WhcNMTUwNDA0MTUxNTU1WjBJMQswCQYDVQQG
EwJVUzETMBEGA1UEChMKR29vZ2xlIEluYzElMCMGA1UEAxMcR29vZ2xlIEludGVy
bmV0IEF1dGhvcml0eSBHMjCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB
AJwqBHdc2FCROgajguDYUEi8iT/xGXAaiEZ+4I/F8YnOIe5a/mENtzJEiaB0C1NP
VaTOgmKV7utZX8bhBYASxF6UP7xbSDj0U/ck5vuR6RXEz/RTDfRK/J9U3n2+oGtv
h8DQUB8oMANA2ghzUWx//zo8pzcGjr1LEQTrfSTe5vn8MXH7lNVg8y5Kr0LSy+rE
ahqyzFPdFUuLH8gZYR/Nnag+YyuENWllhMgZxUYi+FOVvuOAShDGKuy6lyARxzmZ
EASg8GF6lSWMTlJ14rbtCMoU/M4iarNOz0YDl5cDfsCx3nuvRTPPuj5xt970JSXC
DTWJnZ37DhF5iR43xa+OcmkCAwEAAaOB+zCB+DAfBgNVHSMEGDAWgBTAephojYn7
qwVkDBF9qn1luMrMTjAdBgNVHQ4EFgQUSt0GFhu89mi1dvWBtrtiGrpagS8wEgYD
VR0TAQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAQYwOgYDVR0fBDMwMTAvoC2g
K4YpaHR0cDovL2NybC5nZW90cnVzdC5jb20vY3Jscy9ndGdsb2JhbC5jcmwwPQYI
KwYBBQUHAQEEMTAvMC0GCCsGAQUFBzABhiFodHRwOi8vZ3RnbG9iYWwtb2NzcC5n
ZW90cnVzdC5jb20wFwYDVR0gBBAwDjAMBgorBgEEAdZ5AgUBMA0GCSqGSIb3DQEB
BQUAA4IBAQA21waAESetKhSbOHezI6B1WLuxfoNCunLaHtiONgaX4PCVOzf9G0JY
/iLIa704XtE7JW4S615ndkZAkNoUyHgN7ZVm2o6Gb4ChulYylYbc3GrKBIxbf/a/
zG+FA1jDaFETzf3I93k9mTXwVqO94FntT0QJo544evZG0R0SnU++0ED8Vf4GXjza
HFa9llF7b1cq26KqltyMdMKVvvBulRP/F/A8rLIQjcxz++iPAsbw+zOzlTvjwsto
WHPbqCRiOwY1nQ2pM714A5AuTHhdUDqB1O6gyHA43LL5Z/qHQF1hwFGPa4NrzQU6
yuGnBXj8ytqU0CwIPX4WecigUCAkVDNx
-----END CERTIFICATE-----`

func TestParseCertificate(t *testing.T) {
	c, err := parseCertificate(pemData)
	if err != nil {
		t.Error(err)
	}
	if c.Issuer.CommonName != "Google Internet Authority G2" {
		t.Errorf("Wrong organization %s", c.Issuer.CommonName)
	}
	
}

func helperCertificateRequest(
	t *testing.T,
	f func() ([]*certificate, error),
	casseteName string,
	c *SSLCollector,
	numbCerts int,
	clientShouldSuceed bool) {

		vcr := govcr.NewVCR(
			casseteName,
			&govcr.VCRConfig{
				Client:    c.httpClient,
				RemoveTLS: true,
		})

		c.httpClient = vcr.Client
	
		certs, err := f()
		if clientShouldSuceed && err != nil {
			t.Error(err)
		}
		if !clientShouldSuceed && err == nil {
			t.Error(errors.New("there should have been an error"))
		}
		if len(certs) != numbCerts {
			t.Errorf("Wrong number of certs, %d should be %d", len(certs), numbCerts)
		}
		fmt.Printf("govcr stats %+v\n", vcr.Stats())
}

func TestFetchFromCloudSQL(t *testing.T) {
	client, err := getHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	c := NewSSLCollector([]string{"sojern-dev"}, client, false)
	helperCertificateRequest(
		t,
		c.fetchFromCloudSQL,
		"request_cloudsql_certificates",
		c,
		2, true)
}

func TestFetchFromComputeOnlyInUse(t *testing.T) {
	client, err := getHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	c := NewSSLCollector([]string{"sojern-dev"}, client, true)
	helperCertificateRequest(
		t,
		c.fetchFromCompute,
		"request_compute_certificates_only_in_use",
		c,
		2, true)
}

func TestFetchFromCompute(t *testing.T) {
	client, err := getHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	c := NewSSLCollector([]string{"sojern-platform", "sojern-sre-prod"}, client, false)
	helperCertificateRequest(
		t,
		c.fetchFromCompute,
		"request_compute_certificates",
		c,
		14, true)
}

func TestFetchFromGCPMultipleProjects(t *testing.T) {
	client, err := getHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	c := NewSSLCollector([]string{"sojern-platform", "sojern-sre-prod"}, client, false)
	helperCertificateRequest(
		t,
		c.fetchFromGCP,
		"request_certificates_to_multiple_projects_gcp",
		c,
		16, true)
}

func TestFetchFromGCPUnexistentProjects(t *testing.T) {
	client, err := getHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	c := NewSSLCollector([]string{"sojern-unexistent-project"}, client, false)
	helperCertificateRequest(
		t,
		c.fetchFromGCP,
		"request_certificates_to_unexistent_project",
		c,
		0, false)
}

func TestToInternalCertificates(t *testing.T) {
	numberOfCerts := 5
	projectName   := "project-name"
	dummyCertName := "mycert"
	var Certs []*gcpCertificate

	for i := 0; i < numberOfCerts; i++ {
		Certs = append(Certs, &gcpCertificate{name: dummyCertName, raw: pemData})
	}
	certs, err := toInternalCertificates(Certs, projectName)
	if err != nil {
		t.Error(err)
	}
	if len(certs) != numberOfCerts {
		t.Errorf("Wrong number of certs %d", len(certs))
	}
	for _, c := range certs {
		if c.secondsToExpire == 0.0 || c.name != dummyCertName || c.project != projectName {
			t.Errorf("The following certificate struc is wrong %#v", c)
		}
	}
} 

func TestCollect(t *testing.T) {
	ch := make(chan prometheus.Metric)

	client, err := getHTTPClient()
	if err != nil {
		t.Fatal(err)
	}
	vcr := govcr.NewVCR("prometheus_collect",
		&govcr.VCRConfig{
			Client:    client,
			RemoveTLS: true,
	})
	collector := NewSSLCollector(
		[]string{"sojern-platform", "sojern-dev"}, 
		vcr.Client,
		false)

	go func() {
		collector.Collect(ch)
	}()
	<-ch
	fmt.Printf("govcr stats %+v\n", vcr.Stats())
}

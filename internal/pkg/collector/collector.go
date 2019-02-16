package collector

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	c "github.com/snebel29/prometheus-gcp-ssl-exporter/internal/pkg/cli"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/sqladmin/v1beta4"
)

// Run the collector to accept http scraping requests
func Run(cli *c.CLI) error {
	client, err := getHTTPClient()
	if err != nil {
		return err
	}
	Register(cli.Projects, client, cli.OnlyInUse)
	http.Handle(cli.MetricsPath, promhttp.Handler())
	log.Infof("Beginning to serve on port :%s", cli.Port)
	return http.ListenAndServe(fmt.Sprintf(":%s", cli.Port), nil)
}

// Register instantiate as new SSL collector then registers with prometheus
func Register(projects []string, client *http.Client, onlyInUse bool) {
	prometheus.MustRegister(NewSSLCollector(projects, client, onlyInUse))
}

// SSLCollector represents the collector
type SSLCollector struct {
	sslValidity *prometheus.Desc
	projects    []string
	httpClient  *http.Client
	onlyInUse   bool   // Whether we should fetch compute certs in use by httpsProxies only
}

// NewSSLCollector Returns a new ssl collector
func NewSSLCollector(projects []string, client *http.Client, onlyInUse bool) *SSLCollector {
	variableLabels := []string{"name", "project", "service"}
	return &SSLCollector{
		sslValidity: prometheus.NewDesc("gcp_ssl_validity_seconds",
			"Time for an ssl certificate to expire",
			variableLabels, nil),
		projects:   projects,
		httpClient: client,
		onlyInUse:  onlyInUse,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
func (c *SSLCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.sslValidity
}

// Collect is called by the Prometheus registry when collecting metrics
func (c *SSLCollector) Collect(ch chan<- prometheus.Metric) {
	valueList, err := c.fetchFromGCP()
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	for _, v := range valueList {
		metric, err := prometheus.NewConstMetric(
			c.sslValidity,
			prometheus.GaugeValue,
			v.secondsToExpire,
			v.name,
			v.project,
			v.service,
		)

		if err != nil {
			log.Errorf("%s", err)
		} else {
			ch <- metric
		}
	}
}

type gcpCertificate struct {
	name    string
	raw     string
	service string
}

type certificate struct {
	name            string
	project         string
	service         string
	secondsToExpire float64
}

func getHTTPClient() (*http.Client, error) {
	c, err := google.DefaultClient(context.Background(), "")
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *SSLCollector) fetchFromGCP() ([]*certificate, error) {
	comp, err := c.fetchFromCompute()
	if err != nil {
		return nil, err
	}
	cloud, err := c.fetchFromCloudSQL()
	if err != nil {
		return nil, err
	}
	combined := append(comp, cloud...)
	return combined, nil
}

func (c *SSLCollector) fetchFromCloudSQL() ([]*certificate, error) {
	svc, err := sqladmin.New(c.httpClient)
	if err != nil {
		e := fmt.Sprintf("Trying to instantiate cloudsql service: [%s]", err)
		return nil, errors.New(e)
	}

	var projectsCertificates []*certificate

	for _, project := range c.projects {
		instances, err := svc.Instances.List(project).Do()
		if err != nil {
			e := fmt.Sprintf("Trying to list instances for instance project [%s] with error [%s]", project, err)
			return nil, errors.New(e)
		}
		for _, instance := range instances.Items {

			certificates, err := svc.SslCerts.List(project, instance.Name).Do()
			// TODO: Return data from successfull projects in partial failures scenarios
			if err != nil {
				e := fmt.Sprintf("Trying to list certificates for instance [%s] in project [%s] with error [%s]", instance.Name, project, err)
				return nil, errors.New(e)
			}

			certs, err := toInternalCertificates(getCertificateFromCloudsqlAPICertificate(certificates), project)
			if err != nil {
				return nil, err
			}

			projectsCertificates = append(projectsCertificates, certs...)
		}
	}
	return projectsCertificates, nil
}

// Fetch certificates from compute API which are bind to an httpsProxy
func (c *SSLCollector) fetchFromComputeOnlyInUse(svc *compute.Service) ([]*certificate, error) {
	var projectsCertificates []*certificate

	for _, project := range c.projects {
		httpsProxies, err := svc.TargetHttpsProxies.List(project).Do()
		// TODO: Return data from successfull projects in partial failures scenarios
		if err != nil {
			e := fmt.Sprintf("Trying to list httpsProxies in project [%s] with error [%s]", project, err)
			return nil, errors.New(e)
		}

		var m map[string]bool
		m = make(map[string]bool)

		for _, httpsProxy := range httpsProxies.Items {


			for _, httpsProxyCertURI := range httpsProxy.SslCertificates {
				s := strings.Split(httpsProxyCertURI, "/")
				httpsProxyCertName := s[len(s)-1]

				// Same certificate could be bind to multiple httpsProxies, we don't want duplicates
				if _, exists := m[httpsProxyCertName]; exists { break }
				m[httpsProxyCertName] = true

				hc, err := svc.SslCertificates.Get(project, httpsProxyCertName).Do()
				if err != nil {
					e := fmt.Sprintf("Trying to get certificate [%s] in project [%s] with error [%s]", httpsProxyCertName, project, err)
					return nil, errors.New(e)
				}

				certificates := &compute.SslCertificateList{Items: []*compute.SslCertificate{hc}}
				certs, err := toInternalCertificates(getCertificateFromComputeAPICertificate(certificates), project)
				if err != nil {
					return nil, err
				}
				projectsCertificates = append(projectsCertificates, certs...)
			}
		}
	}

	return projectsCertificates, nil
}

// Fetch all certificates from compute API even if they are not bind to an httpsProxy
func (c *SSLCollector) fetchFromComputeAll(svc *compute.Service) ([]*certificate, error) {
	var projectsCertificates []*certificate

	for _, project := range c.projects {
		certificates, err := svc.SslCertificates.List(project).Do()
		// TODO: Return data from successfull projects in partial failures scenarios
		if err != nil {
			e := fmt.Sprintf("Trying to list certificates in project [%s] with error [%s]", project, err)
			return nil, errors.New(e)
		}

		certs, err := toInternalCertificates(getCertificateFromComputeAPICertificate(certificates), project)
		if err != nil {
			return nil, err
		}
		projectsCertificates = append(projectsCertificates, certs...)
	}

	return projectsCertificates, nil
}

func (c *SSLCollector) fetchFromCompute() ([]*certificate, error) {
	svc, err := compute.New(c.httpClient)
	if err != nil {
		e := fmt.Sprintf("Trying to instantiate compute service: [%s]", err)
		return nil, errors.New(e)
	}

	var f func(svc *compute.Service) ([]*certificate, error)

	if c.onlyInUse {
		f = c.fetchFromComputeOnlyInUse
	} else {
		f = c.fetchFromComputeAll
	}

	projectsCertificates, err := f(svc)
	if err != nil {
		e := fmt.Sprintf("Trying to fetch from compute service: [%s]", err)
		return nil, errors.New(e)	
	}
	return projectsCertificates, nil
}

func getCertificateFromComputeAPICertificate(certs *compute.SslCertificateList) []*gcpCertificate {
	var gcpCerts []*gcpCertificate
	for _, c := range certs.Items {
		gcpCerts = append(gcpCerts, &gcpCertificate{
			name:    c.Name,
			raw:     c.Certificate,
			service: "compute",
		})
	}
	return gcpCerts
}

func getCertificateFromCloudsqlAPICertificate(certs *sqladmin.SslCertsListResponse) []*gcpCertificate {
	var gcpCerts []*gcpCertificate
	for _, c := range certs.Items {
		gcpCerts = append(gcpCerts, &gcpCertificate{
			name:    fmt.Sprintf("%s-%s", c.Instance, c.CommonName),
			raw:     c.Cert,
			service: "cloudsql",
		})
	}
	return gcpCerts
}

func toInternalCertificates(gcpCertList []*gcpCertificate, project string) ([]*certificate, error) {

	var projectsCertificates []*certificate
	for _, cert := range gcpCertList {
		c, err := parseCertificate(cert.raw)
		if err != nil {
			return nil, err
		}
		secondsToExpire := float64(c.NotAfter.Unix() - time.Now().Unix())
		log.Debugf("%v %v %v %v %v",
			cert.name, c.NotAfter, time.Now().Unix(), secondsToExpire, c.NotAfter.Unix())

		projectsCertificates = append(projectsCertificates, &certificate{
			name:            cert.name,
			project:         project,
			secondsToExpire: secondsToExpire,
			service:         cert.service})
	}
	return projectsCertificates, nil
}

func parseCertificate(raw string) (*x509.Certificate, error) {
	var nilCertificate *x509.Certificate
	var blocks []byte
	remainder := []byte(raw)
	for {
		var block *pem.Block
		block, remainder = pem.Decode(remainder)
		if block == nil {
			return nilCertificate, errors.New("PEM not parsed")
		}
		blocks = append(blocks, block.Bytes...)
		if len(remainder) == 0 {
			break
		}
	}
	c, err := x509.ParseCertificates(blocks)
	if err != nil {
		return nilCertificate, err
	}
	return c[0], nil
}

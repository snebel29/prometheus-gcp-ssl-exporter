package collector

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	c "github.com/snebel29/prometheus-gcp-ssl-exporter/internal/pkg/cli"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

// Run the collector to accept http scraping requests
func Run(cli *c.CLI) error {
	client, err := getHTTPClient()
	if err != nil {
		return err
	}
	Register(cli.Projects, client)
	http.Handle(cli.MetricsPath, promhttp.Handler())
	log.Infof("Beginning to serve on port :%s", cli.Port)
	return http.ListenAndServe(fmt.Sprintf(":%s", cli.Port), nil)
}

// Register instantiate as new SSL collector then registers with prometheus
func Register(projects []string, client *http.Client) {
	prometheus.MustRegister(NewSSLCollector(projects, client))
}

// SSLCollector represents the collector
type SSLCollector struct {
	sslValidity *prometheus.Desc
	projects    []string
	httpClient  *http.Client
}

// NewSSLCollector Returns a new ssl collector
func NewSSLCollector(projects []string, client *http.Client) *SSLCollector {
	variableLabels := []string{"name", "project"}
	return &SSLCollector{
		sslValidity: prometheus.NewDesc("gcp_ssl_validity_seconds",
			"Time for an ssl certificate to expire",
			variableLabels, nil),
		projects: projects,
		httpClient: client,
	}
}

// Describe sends the super-set of all possible descriptors of metrics
func (collector *SSLCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.sslValidity
}

// Collect is called by the Prometheus registry when collecting metrics
func (collector *SSLCollector) Collect(ch chan<- prometheus.Metric) {
	valueList, err := fetchFromGCP(collector.projects, collector.httpClient)
	if err != nil {
		log.Errorf("%s", err)
		return
	}

	for _, c := range valueList {
		metric, err := prometheus.NewConstMetric(
			collector.sslValidity,
			prometheus.GaugeValue,
			c.secondsToExpire,
			c.name,
			c.project,
		)
	
		if err != nil {
			log.Errorf("%s", err)
		} else {
			ch <- metric
		}
	}
}

type certificate struct {
	name 			string
	project         string
	secondsToExpire float64
}

func getHTTPClient() (*http.Client, error) {
	c, err := google.DefaultClient(context.Background(), compute.ComputeScope)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func fetchFromGCP(projects []string, client *http.Client) ([]*certificate, error) {
	svc, err := compute.New(client)
	if err != nil {
		e := fmt.Sprintf("Trying to instantiate compute service: [%s]", err)
		return nil, errors.New(e)
	}

	var projectsCertificates []*certificate

	for _, project := range projects {
		certificates, err := svc.SslCertificates.List(project).Do()
		// TODO: Return data from successfull projects in partial failures scenarios
		if err != nil {
			e := fmt.Sprintf("Trying to list certificates in project [%s] with error [%s]", project, err)
			return nil, errors.New(e)
		}

		c, err := toInternalCertificates(certificates, project)
		if err != nil {
			return nil, err
		}

		projectsCertificates = append(projectsCertificates, c...)
	}
	return projectsCertificates, nil
}

func toInternalCertificates(certList *compute.SslCertificateList, project string) ([]*certificate, error) {

	var projectsCertificates []*certificate
	for _, cert := range certList.Items {
		c, err := parseCertificate(cert.Certificate)
		if err != nil {
			return nil, err
		}
		secondsToExpire := float64(c.NotAfter.Unix() - time.Now().Unix())
		log.Debugf("%v %v %v %v %v", 
			cert.Name, c.NotAfter, time.Now().Unix(), secondsToExpire, c.NotAfter.Unix())

		projectsCertificates = append(projectsCertificates, &certificate{
													name: cert.Name,
													project: project,
													secondsToExpire: secondsToExpire})							
	}
	return projectsCertificates, nil
}

func parseCertificate(cert string) (*x509.Certificate, error) {
	var nilCertificate *x509.Certificate
	var blocks []byte
	remainder := []byte(cert)
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

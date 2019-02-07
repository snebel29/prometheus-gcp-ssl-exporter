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
	"google.golang.org/api/sqladmin/v1beta4"
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

type gcpCertificate struct {
	name string
	raw  string
}

type certificate struct {
	name 			string
	project         string
	secondsToExpire float64
}

func getHTTPClient() (*http.Client, error) {
	c, err := google.DefaultClient(context.Background(), "")
	if err != nil {
		return nil, err
	}
	return c, nil
}

func fetchFromCloudSQL(projects []string, client *http.Client) ([]*certificate, error) {
	svc, err := sqladmin.New(client)
	if err != nil {
		e := fmt.Sprintf("Trying to instantiate cloudsql service: [%s]", err)
		return nil, errors.New(e)
	}

	var projectsCertificates []*certificate

	for _, project := range projects {
		instances, err := svc.Instances.List(project).Do()
		if err != nil {
			e := fmt.Sprintf("Trying to list instnces for instance project [%s] with error [%s]", project, err)
			return nil, errors.New(e)
		}
		for _, instance := range instances.Items {

			certificates, err := svc.SslCerts.List(project, instance.Name).Do()
			// TODO: Return data from successfull projects in partial failures scenarios
			if err != nil {
				e := fmt.Sprintf("Trying to list certificates for instance [%s] in project [%s] with error [%s]", instance.Name, project, err)
				return nil, errors.New(e)
			}

			c, err := toInternalCertificates(getCertificateFromCloudsqlAPICertificate(certificates), project)
			if err != nil {
				return nil, err
			}

			projectsCertificates = append(projectsCertificates, c...)
		}
	}
	return projectsCertificates, nil
}

func fetchFromCompute(projects []string, client *http.Client) ([]*certificate, error) {
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

		c, err := toInternalCertificates(getCertificateFromComputeAPICertificate(certificates), project)
		if err != nil {
			return nil, err
		}

		projectsCertificates = append(projectsCertificates, c...)
	}
	return projectsCertificates, nil
}

func fetchFromGCP(projects []string, client *http.Client) ([]*certificate, error) {
	comp, err := fetchFromCompute(projects, client)
	if err != nil {
		return nil, err
	}
	cloud, err := fetchFromCloudSQL(projects, client)
	if err != nil {
		return nil, err
	}
	combined := append(comp, cloud...)
	return combined, nil
}

func getCertificateFromComputeAPICertificate(certs *compute.SslCertificateList) []*gcpCertificate {
	var gcpCerts []*gcpCertificate
	for _, c := range certs.Items {
		gcpCerts = append(gcpCerts, &gcpCertificate{
			name: c.Name,
			raw:  c.Certificate,
		})
	}
	return gcpCerts
}

func getCertificateFromCloudsqlAPICertificate(certs *sqladmin.SslCertsListResponse) []*gcpCertificate {
	var gcpCerts []*gcpCertificate
	for _, c := range certs.Items {
		gcpCerts = append(gcpCerts, &gcpCertificate{
			name: fmt.Sprintf("%s-%s", c.Instance, c.CommonName),
			raw:  c.Cert,
		})
	}
	return gcpCerts
}

func toInternalCertificates(certList []*gcpCertificate, project string) ([]*certificate, error) {

	var projectsCertificates []*certificate
	for _, cert := range certList {
		c, err := parseCertificate(cert.raw)
		if err != nil {
			return nil, err
		}
		secondsToExpire := float64(c.NotAfter.Unix() - time.Now().Unix())
		log.Debugf("%v %v %v %v %v", 
			cert.name, c.NotAfter, time.Now().Unix(), secondsToExpire, c.NotAfter.Unix())

		projectsCertificates = append(projectsCertificates, &certificate{
													name: cert.name,
													project: project,
													secondsToExpire: secondsToExpire})							
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

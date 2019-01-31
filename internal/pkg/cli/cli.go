package cli

import "gopkg.in/alecthomas/kingpin.v2"

var (
	// Version of the exporter to be set through linker ldflags
	Version string
	metricsPath = kingpin.Flag(
		"metrics-path", "URI path where metrics will be exposed").Default("/metrics").Short('m').String()
	port = kingpin.Flag(
		"port", "Port to listen on").Default("8888").String()
	project = kingpin.Flag(
		"project", "GCP project where to fetch certificates from").Required().Short('p').Strings()
)

// CLI holds command line arguments
type CLI struct {
	MetricsPath string
	Port 	    string
	Projects    []string
}

// NewCLI returns a CLI
func NewCLI() *CLI {
	kingpin.Version(Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	return &CLI{
		MetricsPath: *metricsPath,
		Port:	     *port,
		Projects:    *project,
	}
}
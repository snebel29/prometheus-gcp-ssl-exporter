package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/snebel29/prometheus-gcp-ssl-exporter/internal/pkg/cli"
	"github.com/snebel29/prometheus-gcp-ssl-exporter/internal/pkg/collector"
)

func main() {
	log.Fatal(collector.Run(cli.NewCLI()))
}

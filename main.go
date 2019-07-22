// splunk_exporter
// Prometheus exporter for Splunk API data
//
// Copyright 2019, MongoDB, Inc.
//
// All rights reserved - Do Not Redistribute

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/ehershey/splunk-exporter/handler"
	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile    = kingpin.Flag("config", "Config File").Short('c').Default(fmt.Sprintf("%s.yml", path.Base(os.Args[0]))).String()
	localCertFile = kingpin.Flag("localCert", "Splunk x509 cert File").Short('x').String()
)

func main() {
	// Process command-line arguments
	kingpin.CommandLine.HelpFlag.Short('h')

	kingpin.Parse()

	// Process config
	cfg := newDefaultConfig()

	configContent, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Can't load config: %v", err)
	}

	err = yaml.Unmarshal([]byte(configContent), &cfg)
	if err != nil {
		log.Fatalf("Can't load config: %v", err)
	}

	// Set up the logger
	//if err := logging.Initialize(cfg.Logging); err != nil {
	//log.Fatalf("Can't initialize logging: %v", err)
	//}
	defer zap.L().Sync()

	// build an http client

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	// Read in a custom cert file
	if *localCertFile != "" {
		certs, err := ioutil.ReadFile(*localCertFile)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", localCertFile, err)
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("Cert append failed, using system certs only")
		}
	}
	// Trust the augmented cert pool in our client
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: config}
	client := &http.Client{Transport: tr}

	// BasicAuth: [ cfg.username, cfg.password ]
	// }

	// Set up prom handler for splunk data
	splunkHandler, err := handler.NewSplunkHandler(client, cfg.Username, cfg.Password)
	if err != nil {
		zap.S().Fatalw("Can't initialize splunk handler.",
			"err", err)
	}

	zap.S().Info("Splunk exporter started...")

	// Expose the exporter's own metrics
	http.Handle(cfg.MetricsPath, promhttp.Handler())

	// Expose the Splunk metrics
	http.Handle(cfg.ScrapePath, splunkHandler)

	err = http.ListenAndServe(cfg.ListenAddress, nil)
	if err != nil {
		zap.S().Fatalw("HTTP ListenAndServe() error:",
			"err", err,
		)
	}
}

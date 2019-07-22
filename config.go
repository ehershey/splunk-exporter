package main

import ()

type config struct {
	ListenAddress string `yaml:"listen_address,omitempty"`
	MetricsPath   string `yaml:"metrics_path,omitempty"`
	ScrapePath    string `yaml:"scrape_path,omitempty"`
	Password      string `yaml:"password,omitempty"`
	Username      string `yaml:"username,omitempty"`
	// Logging       logging.Config `yaml:"logging,omitempty"`
}

func newDefaultConfig() *config {
	return &config{
		ListenAddress: ":9042",
		MetricsPath:   "/metrics",
		ScrapePath:    "/scrape",
	}
}

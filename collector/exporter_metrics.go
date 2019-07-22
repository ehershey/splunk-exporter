package collector

import "github.com/prometheus/client_golang/prometheus"

type ExporterMetrics struct {
	TotalApiRequests   *prometheus.CounterVec
	RequestTimeSummary prometheus.Summary
	ScrapeTimeSummary  prometheus.Summary
}

// NewExporterMetrics creates and registers exporter meta-metrics with the global registerer
func NewExporterMetrics() (*ExporterMetrics, error) {
	totalApiRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "splunk_api_calls_total",
			Help: "API requests made to Splunk",
		},
		[]string{"status"},
	)
	if err := prometheus.Register(totalApiRequests); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			totalApiRequests = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			return nil, err
		}
	}

	requestTimeSummary := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "splunk_api_request_time",
			Help:       "API request time",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
	)
	if err := prometheus.Register(requestTimeSummary); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			requestTimeSummary = are.ExistingCollector.(prometheus.Summary)
		} else {
			return nil, err
		}
	}

	scrapeTimeSummary := prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "splunk_scrape_time",
			Help:       "Total Splunk scrape request time",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
	)
	if err := prometheus.Register(scrapeTimeSummary); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			scrapeTimeSummary = are.ExistingCollector.(prometheus.Summary)
		} else {
			return nil, err
		}
	}

	return &ExporterMetrics{
		totalApiRequests,
		requestTimeSummary,
		scrapeTimeSummary,
	}, nil
}

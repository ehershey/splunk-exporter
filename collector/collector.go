package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	typeFloat64 = reflect.TypeOf(float64(0))
	typeBool    = reflect.TypeOf(true)
)

// SplunkCollector implements the Prometheus Collector interface
type SplunkCollector struct {
	metrics []prometheus.Metric
}

// NewSplunkCollector will create a new SplunkCollector and scrape the passed Node attributes, exposing the results
// as Prometheus metrics.
// NewSplunkCollector will also gather metrics on the scrape itself
func NewSplunkCollector(client *http.Client, username string, password string, rm *ExporterMetrics) (*SplunkCollector, error) {
	c := &SplunkCollector{
		metrics: make([]prometheus.Metric, 0),
	}

	// Get the data from the Splunk REST API
	rm.TotalApiRequests.With(prometheus.Labels{"status": "attempted"}).Inc()
	apiNow := time.Now()
	url := "https://localhost:8089/services/server/health/splunkd/details?output_mode=json"
	req, err := http.NewRequest("GET", url, nil)

	req.SetBasicAuth(username, password)

	if err != nil {
		rm.TotalApiRequests.With(prometheus.Labels{"status": "errored"}).Inc()
		return nil, errors.Wrap(err, fmt.Sprintf("Error creating http request to scrape URL: %v", url))
	}

	resp, err := client.Do(req)
	if err != nil {
		rm.TotalApiRequests.With(prometheus.Labels{"status": "errored"}).Inc()
		return nil, errors.Wrap(err, fmt.Sprintf("Error scraping URL: %v", url))
	}
	rm.RequestTimeSummary.Observe(time.Since(apiNow).Seconds())
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Error from Splunk: %v", string(body))
	}

	rm.TotalApiRequests.With(prometheus.Labels{"status": "succeeded"}).Inc()

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error parsing json from URL: %v", url))
	}
	log.Println("dat:", dat)

	entryList := dat["entry"].([]interface{})

	log.Println("entryList:", entryList)

	healthEntry := entryList[0].(map[string]interface{})

	log.Println("healthEntry:", healthEntry)

	healthContent := healthEntry["content"].(map[string]interface{})

	log.Println("healthContent:", healthContent)

	serverHealth := healthContent["health"]

	log.Println("serverHealth:", serverHealth)

	var floatVal float64
	if serverHealth == "green" {
		floatVal = 1.0
	} else if serverHealth == "yellow" {
		floatVal = 0.5
	} else if serverHealth == "red" {
		floatVal = 0.0
	} else {
		return nil, fmt.Errorf("Unknown health string in Splunk response: %v", serverHealth)
	}

	// // Set the metrics labels
	// // in addition to the configured labels, each metric
	// // as the instance_id and attribute as a label
	// labels := make(map[string]string, 0)
	// if a.Labels != nil {
	// for k, v := range a.Labels {
	// labels[k] = v
	// }
	// }
	// labels["instance_id"] = target
	// labels["attribute"] = strings.Join(a.Path, ".")

	desc := prometheus.NewDesc(
		"splunk_health",
		"Splunk process health",
		nil,
		nil,
	)

	c.metrics = append(
		c.metrics,
		prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			floatVal,
		),
	)
	return c, nil
}

func (c *SplunkCollector) Collect(ch chan<- prometheus.Metric) {
	for m := range c.metrics {
		ch <- c.metrics[m]
	}
}

func (c *SplunkCollector) Describe(ch chan<- *prometheus.Desc) {
	for m := range c.metrics {
		ch <- c.metrics[m].Desc()
	}
}

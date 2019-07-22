package handler

import (
	"net/http"
	"time"

	"github.com/ehershey/splunk-exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type splunkHandler struct {
	c        *http.Client
	username string
	password string
	e        *collector.ExporterMetrics
}

func NewSplunkHandler(c *http.Client, username string, password string) (http.Handler, error) {
	e, err := collector.NewExporterMetrics()
	if err != nil {
		return nil, err
	}

	return &splunkHandler{
		c:        c,
		username: username,
		password: password,
		e:        e,
	}, nil
}

func (h *splunkHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	now := time.Now()

	// scrape the requested data from Splunk
	col, err := collector.NewSplunkCollector(h.c, h.username, h.password, h.e)
	if err != nil {
		http.Error(w, err.Error(), 500)
		instrumentMetricsError("500")
		return
	}

	// create a promhttp handler, register the collected metrics, and serve them
	r := prometheus.NewRegistry()
	r.MustRegister(col)

	ha := promhttp.HandlerFor(r, promhttp.HandlerOpts{
		DisableCompression: false,
	})
	ha = promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, ha)

	ha.ServeHTTP(w, req)
	h.e.ScrapeTimeSummary.Observe(time.Since(now).Seconds())
}

// instrumentMetricsError reports errors to the promhttp_metric_handler_requests_total metric.
// This happens even if the request fails before being registered with the promhttp handler.
func instrumentMetricsError(code string) {
	/*
	 * If the handler bails out before passing off to the promhttp handler,
	 * we should still pass the error information to the promhttp metric:
	 * promhttp_metric_handler_requests_total
	 */

	cnt := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "promhttp_metric_handler_requests_total",
			Help: "Total number of scrapes by HTTP status code.",
		},
		[]string{"code"},
	)
	if err := prometheus.DefaultRegisterer.Register(cnt); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			cnt = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			panic(err)
		}
	}

	cnt.With(prometheus.Labels{"code": code}).Inc()
}

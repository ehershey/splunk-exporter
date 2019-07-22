# splunk-exporter
Prometheus exporter for Splunk

## Building
go build
## Running
./splunk-exporter --config configs/splunk-exporter.yml
open http://localhost:9042/scrape
open http://localhost:9042/metrics

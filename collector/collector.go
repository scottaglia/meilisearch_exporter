package collector

import (
	"net/url"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/meilisearch/meilisearch-go"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "meilisearch"

var (
	defaultIndicesTotalFieldsLabels = []string{"index"}

	numberOfDocuments = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "number_of_documents"),
		"Total number of documents",
		defaultIndicesTotalFieldsLabels,
		nil,
	)
	isIndexing = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "is_indexing"),
		"If 1, the index is still processing documents and attempts to search will result in undefined behavior. If 0, the index has finished processing and you can start searching",
		defaultIndicesTotalFieldsLabels,
		nil,
	)
)

type MeilisearchCollector struct {
	logger   log.Logger
	msURL    *url.URL
	msClient *meilisearch.Client

	up, lastUpdate, databaseSize prometheus.Gauge
	totalScrapes                 prometheus.Counter
}

func NewCollector(logger log.Logger, timeout *time.Duration, msURL *url.URL, msApiKey *string) *MeilisearchCollector {
	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    msURL.String(),
		Timeout: *timeout,
		APIKey:  *msApiKey,
	})

	return &MeilisearchCollector{
		logger:   logger,
		msURL:    msURL,
		msClient: client,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "", "up"),
			Help: "Meilisearch instance is up and running",
		}),

		lastUpdate: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "", "last_update"),
			Help: "When the last update was made to the database",
		}),

		databaseSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "", "database_size"),
			Help: "Size of the database in bytes",
		}),

		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "", "total_scrapes"),
			Help: "Total of Meilisearch scrapes.",
		}),
	}
}

func (mc *MeilisearchCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- mc.up.Desc()
	ch <- mc.totalScrapes.Desc()
}

func (mc *MeilisearchCollector) Collect(ch chan<- prometheus.Metric) {
	mc.totalScrapes.Inc()

	defer func() {
		ch <- mc.totalScrapes
		ch <- mc.up
		ch <- mc.lastUpdate
		ch <- mc.databaseSize
	}()

	if mc.msClient.IsHealthy() {
		mc.up.Set(1)
	} else {
		mc.up.Set(0)
	}

	stats, err := mc.msClient.GetStats()
	if err != nil {
		level.Error(mc.logger).Log("msg", "failed to fetch statistics", "err", err)
		return
	}

	last_update := stats.LastUpdate.Unix()
	if stats.LastUpdate.IsZero() {
		last_update = 0
	}

	mc.lastUpdate.Set(float64(last_update))
	mc.databaseSize.Set(float64(stats.DatabaseSize))

	for indexName, indexStats := range stats.Indexes {
		is_indexing := 0
		if indexStats.IsIndexing {
			is_indexing = 1
		}

		ch <- prometheus.MustNewConstMetric(
			numberOfDocuments,
			prometheus.GaugeValue,
			float64(indexStats.NumberOfDocuments),
			indexName,
		)

		ch <- prometheus.MustNewConstMetric(
			isIndexing,
			prometheus.GaugeValue,
			float64(is_indexing),
			indexName,
		)
	}

}

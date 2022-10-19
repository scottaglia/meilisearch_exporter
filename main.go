package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/scottaglia/meilisearch_exporter/collector"
	"gopkg.in/alecthomas/kingpin.v2"
)

const name = "meilisearch_exporter"
const address = ":9974"

var (
	// Exporter
	webConfig   = webflag.AddFlags(kingpin.CommandLine, address)
	metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()

	// Meilisearch
	msURI     = kingpin.Flag("ms.uri", "HTTP API address of the Meilisearch host.").Default("http://localhost:7700").OverrideDefaultFromEnvar("MEILISEARCH_EXPORTER_URI").String()
	msApiKey  = kingpin.Flag("ms.apikey", "Meilisearch API Key.").OverrideDefaultFromEnvar("MEILISEARCH_EXPORTER_APIKEY").String()
	msTimeout = kingpin.Flag("ms.timeout", "Timeout for trying to get stats from Meilisearch.").Default("5s").Duration()
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print(name))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	if len(*msURI) == 0 {
		level.Error(logger).Log("msg", "ms.uri cannot be empty")
		os.Exit(1)
	}

	msURL, err := url.Parse(*msURI)
	if err != nil {
		level.Error(logger).Log("msg", "error parsing ms.uri", "err", err)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", fmt.Sprintf("Starting %s", name), "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())

	prometheus.MustRegister(version.NewCollector(name))
	prometheus.MustRegister(collector.NewCollector(logger, msTimeout, msURL, msApiKey))

	homePage := []byte(`<html>
  <head><title>Meilisearch Exporter</title></head>
  <body>
  <h1>Meilisearch Exporter</h1>
  <p><a href="` + *metricsPath + `">Metrics</a></p>
  </body>
  </html>`)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	server := &http.Server{}
	mux := http.DefaultServeMux

	mux.Handle(*metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write(homePage)
		if err != nil {
			level.Error(logger).Log("msg", "failed handling writer", "err", err)
		}
	})

	server.Handler = mux

	go func() {
		if err := web.ListenAndServe(server, webConfig, logger); err != nil {
			if fmt.Sprintf("%s", err) != "http: Server closed" {
				level.Error(logger).Log("msg", "failed starting http server", "err", err)
			}
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	_ = level.Info(logger).Log("msg", "shutting down")
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer srvCancel()
	_ = server.Shutdown(srvCtx)
}

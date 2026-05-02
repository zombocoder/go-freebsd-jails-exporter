//go:build freebsd && cgo

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/zombocoder/go-freebsd-jails-exporter/internal/buildinfo"
	"github.com/zombocoder/go-freebsd-jails-exporter/internal/collector"
	"github.com/zombocoder/go-freebsd-jails-exporter/internal/logging"
	"github.com/zombocoder/go-freebsd-jails-exporter/pkg/version"
)

func main() {
	app := kingpin.New("go-freebsd-jails-exporter", "Prometheus exporter for FreeBSD jails.")
	app.Version(fmt.Sprintf("%s (revision: %s)", version.Version, version.Revision))

	listenAddress := app.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").
		Default(":9756").String()
	telemetryPath := app.Flag("web.telemetry-path", "Path under which to expose metrics.").
		Default("/metrics").String()

	logLevel := app.Flag("log.level", "Log level (debug, info, warn, error).").
		Default("info").Enum("debug", "info", "warn", "error")
	logFormat := app.Flag("log.format", "Log format (logfmt, json).").
		Default("logfmt").Enum("logfmt", "json")

	jailsEnabled := app.Flag("collector.jails.enabled", "Enable the jails collector.").
		Default("true").Bool()
	processesEnabled := app.Flag("collector.processes.enabled", "Enable the processes collector.").
		Default("true").Bool()
	racctEnabled := app.Flag("collector.racct.enabled", "Enable the RACCT collector.").
		Default("true").Bool()
	zfsEnabled := app.Flag("collector.zfs.enabled", "Enable the ZFS collector.").
		Default("true").Bool()

	collectorTimeout := app.Flag("collector.timeout", "Timeout for collector scrapes.").
		Default("5s").Duration()
	_ = app.Flag("collector.racct.timeout", "Timeout for RACCT collector (unused, uses collector.timeout).").
		Default("2s").Duration()

	_ = app.Flag("config.file", "Path to configuration file (reserved for future use).").
		Default("").String()

	kingpin.MustParse(app.Parse(os.Args[1:]))

	logger, err := logging.New(*logLevel, *logFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}

	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	logger.Info("starting go-freebsd-jails-exporter",
		"version", version.Version,
		"revision", version.Revision,
		"host", host,
	)

	reg := prometheus.NewRegistry()
	reg.MustRegister(buildinfo.NewCollector())

	var collectors []collector.Collector
	if *jailsEnabled {
		collectors = append(collectors, collector.NewJailsCollector(logger, host))
		logger.Info("collector enabled", "collector", "jails")
	}
	if *processesEnabled {
		collectors = append(collectors, collector.NewProcessesCollector(logger, host))
		logger.Info("collector enabled", "collector", "processes")
	}
	if *racctEnabled {
		collectors = append(collectors, collector.NewRACCTCollector(logger, host))
		logger.Info("collector enabled", "collector", "racct")
	}
	if *zfsEnabled {
		collectors = append(collectors, collector.NewZFSCollector(logger, host))
		logger.Info("collector enabled", "collector", "zfs")
	}

	registry := collector.NewRegistry(logger, *collectorTimeout, collectors...)
	reg.MustRegister(registry)

	mux := http.NewServeMux()

	mux.Handle(*telemetryPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorLog:      nil,
		ErrorHandling: promhttp.ContinueOnError,
		Timeout:       *collectorTimeout + 2*time.Second,
	}))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<html>
<head><title>FreeBSD Jails Exporter</title></head>
<body>
<h1>FreeBSD Jails Exporter</h1>
<p><a href="%s">Metrics</a></p>
<p>Version: %s</p>
</body>
</html>`, *telemetryPath, version.Version)
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	server := &http.Server{
		Addr:         *listenAddress,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("listening", "address", *listenAddress, "path", *telemetryPath)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

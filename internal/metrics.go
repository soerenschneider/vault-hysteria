package internal

import (
	"bytes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

const (
	namespace = "vaultpanic"
)

var (
	HeartbeatMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "heartbeat_seconds",
		Help:      "Heartbeat signal",
	})

	FailedSealCallsMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "vault_seal_failures_total",
		Help:      "Errors sealing the vault",
	})

	SealRequestsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "vault_seal_requests_received_total",
		Help:      "The total amount of requests received",
	})

	SealRequestsIgnored = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "vault_seal_requests_ignored_total",
		Help:      "The total amount of received requests but ignored",
	})

	VaultTokenExpiryTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "token_expiry_seconds",
		Help:      "Token expiry date",
	})

	VaultCommunicationErrors = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "vault_communication_errors_total",
		Help:      "Errors while periodically looking up self",
	})
)

func StartMetricsServer(addr string) {
	log.Info().Msgf("Starting metrics server at %s", addr)
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal().Msgf("Can not start metrics server at %s: %v", addr, err)
	}
}

func WriteMetrics(path string) error {
	log.Info().Msgf("Dumping metrics to %s", path)
	metrics, err := dumpMetrics()
	if err != nil {
		log.Info().Msgf("Error dumping metrics: %v", err)
		return err
	}

	err = ioutil.WriteFile(path, []byte(metrics), 0644)
	if err != nil {
		log.Info().Msgf("Error writing metrics to '%s': %v", path, err)
	}
	return err
}

func dumpMetrics() (string, error) {
	var buf = &bytes.Buffer{}
	enc := expfmt.NewEncoder(buf, expfmt.FmtText)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return "", err
	}

	for _, f := range families {
		if err := enc.Encode(f); err != nil {
			log.Info().Msgf("could not encode metric: %s", err.Error())
		}
	}

	return buf.String(), nil
}

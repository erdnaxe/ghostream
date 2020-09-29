package monitoring

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

// Options holds web package configuration
type Options struct {
	ListenAddress string
}

var (
	// WebViewerServed is the total amount of viewer page served
	WebViewerServed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ghostream_web_viewer_served_total",
		Help: "The total amount of viewer served",
	})

	// WebSessions is the total amount of WebRTC session exchange
	WebSessions = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ghostream_web_sessions_total",
		Help: "The total amount of WebRTC sessions exchanged",
	})

	// WebRTCConnectedSessions is the total amount of WebRTC session exchange
	WebRTCConnectedSessions = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ghostream_webrtc_connected_sessions",
		Help: "The current amount of opened WebRTC sessions",
	})
)

func GetGaugeValue(metric prometheus.Gauge) float64 {
	var m = &dto.Metric{}
	if err := metric.Write(m); err != nil {
		log.Fatal(err)
		return 0
	}
	return m.Gauge.GetValue()
}

// Serve monitoring server that expose prometheus metrics
func Serve(cfg *Options) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	log.Printf("Monitoring HTTP server listening on %s", cfg.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.ListenAddress, mux))
}

package monitoring

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.crans.org/nounous/ghostream/internal/config"
)

var (
	// ViewerServed is the total amount of viewer page served
	ViewerServed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ghostream_viewer_served_total",
		Help: "The total amount of viewer served",
	})
)

// ServeHTTP server that expose prometheus metrics
func ServeHTTP(cfg *config.Config) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	log.Printf("Monitoring listening on http://%s/", cfg.Prometheus.ListenAddress)
	log.Fatal(http.ListenAndServe(cfg.Prometheus.ListenAddress, mux))
}

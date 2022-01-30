package transporthttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHandler struct{}

func (h *MetricsHandler) Path() string {
	return "/_metrics"
}

func (h *MetricsHandler) Register(mux *http.ServeMux) {
	mux.Handle(h.Path(), promhttp.Handler())
}

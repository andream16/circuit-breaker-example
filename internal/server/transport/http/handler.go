package transporthttp

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var isCircuitClosed = true

type (
	CloseCircuitHandler struct{}

	OpenCircuitHandler struct{}

	TripHandler struct{}

	MetricsHandler struct{}
)

func (h *CloseCircuitHandler) Path() string {
	return "/circuit/close"
}

func (h *CloseCircuitHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(h.Path(), func(w http.ResponseWriter, r *http.Request) {
		isCircuitClosed = true
	})
}

func (h *OpenCircuitHandler) Path() string {
	return "/circuit/open"
}

func (h *OpenCircuitHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(h.Path(), func(w http.ResponseWriter, r *http.Request) {
		isCircuitClosed = false
	})
}

func (h *MetricsHandler) Path() string {
	return "/_metrics"
}

func (h *MetricsHandler) Register(mux *http.ServeMux) {
	mux.Handle(h.Path(), promhttp.Handler())
}

func (h *TripHandler) Path() string {
	return "/trip"
}

func (h *TripHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc(h.Path(), h.Trip)
}

func (h *TripHandler) Trip(w http.ResponseWriter, r *http.Request) {
	statusCodes := map[bool]int{
		true:  http.StatusOK,
		false: http.StatusServiceUnavailable,
	}

	statusCode := statusCodes[isCircuitClosed]
	log.Printf("tripper will return status code: %d\n", statusCode)
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
}

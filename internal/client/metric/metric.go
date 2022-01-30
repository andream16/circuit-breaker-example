package metric

import (
	"fmt"
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CircuitStatusLabel  = "status"
	CircuitStatusClosed = "closed"
	TripStatusLabel     = "status"
)

var (
	TripRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trip_requests",
			Help: "Status codes of trip responses",
		},
		[]string{TripStatusLabel},
	)

	CircuitStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_status",
			Help: "Status of the circuit: closed, open, half-open",
		},
		[]string{CircuitStatusLabel},
	)
)

func Register() (err error) {
	defer func() {
		e := recover()
		log.Printf("recovering from prometheus panic: %v\n", e)
		err, ok := e.(error)
		if ok {
			err = fmt.Errorf("prometheus registration panicked: %w", err)
			return
		}
		err = fmt.Errorf("unexpected prometheus panic error: %v", err)
	}()

	for _, c := range []prometheus.Collector{TripRequests, CircuitStatus} {
		if err := prometheus.Register(c); err != nil {
			return fmt.Errorf("unexpected error while registering metrics for: %w", err)
		}
	}

	return nil
}

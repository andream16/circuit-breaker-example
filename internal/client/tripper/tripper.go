package tripper

import (
	"context"
	"fmt"
	"github.com/andream16/circuit-breaker-example/internal/client/metric"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

const tripperName = "cb-tripper"

type (
	Doer interface {
		Do(r *http.Request) (*http.Response, error)
	}

	HttpTripper struct {
		tripURL string
		doer    Doer
		cb      *gobreaker.CircuitBreaker
	}
)

func NewHttpTripper(tripURL string) HttpTripper {
	return HttpTripper{
		tripURL: tripURL,
		doer:    http.DefaultClient,
		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name: tripperName,
			// When to flush counters int the Closed state
			Interval: 15 * time.Second,
			// Time to switch from Open to Half-open
			Timeout: 10 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				log.Printf("total failures %d; requests %d\n", counts.TotalFailures, counts.Requests)
				failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
				log.Printf("failure ratio %f\n", failureRatio)
				return counts.Requests >= 3 && failureRatio >= 0.6
			},
			OnStateChange: func(_ string, from gobreaker.State, to gobreaker.State) {
				metric.CircuitStatus.With(prometheus.Labels{metric.CircuitStatusLabel: to.String()}).Inc()
				// Handler for every state change. We'll use for debugging purpose
				log.Printf("state changed from %q to %q\n", from.String(), to.String())
			},
		}),
	}
}

func (ht HttpTripper) Trip(ctx context.Context) error {
	return ht.trip(ctx)
}

func (ht HttpTripper) trip(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
			if err := ht.do(ctx); err != nil {
				log.Printf("%v\n", err)
			}
		}
	}
}

func (ht HttpTripper) do(ctx context.Context) error {
	if _, err := ht.cb.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, ht.tripURL, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create trip request: %w", err)
		}

		resp, err := ht.doer.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not perform trip request: %w", err)
		}

		defer resp.Body.Close()
		log.Printf("tripper replied with: %d\n", resp.StatusCode)

		metric.TripRequests.With(
			prometheus.Labels{
				metric.TripStatusLabel: fmt.Sprintf("%d", resp.StatusCode),
			},
		).Inc()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		metric.CircuitStatus.With(prometheus.Labels{metric.CircuitStatusLabel: metric.CircuitStatusClosed}).Inc()

		return nil, nil
	}); err != nil {
		return fmt.Errorf("could not execute trip request: %w", err)
	}

	return nil
}

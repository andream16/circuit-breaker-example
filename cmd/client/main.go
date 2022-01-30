package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/andream16/circuit-breaker-example/internal/client/metric"
	clienttransporthttp "github.com/andream16/circuit-breaker-example/internal/client/transport/http"
	"github.com/andream16/circuit-breaker-example/internal/client/tripper"
	transporthttp "github.com/andream16/circuit-breaker-example/internal/transport/http"
)

func main() {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		ht          = tripper.NewHttpTripper("http://server:8080/trip")
	)
	defer cancel()

	if err := metric.Register(); err != nil {
		log.Printf("unexpected metrics registration error: %v\n", err)
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	srv := transporthttp.NewServer(":8080", &clienttransporthttp.MetricsHandler{})

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		select {
		case <-ctx.Done():
		case <-sigs:
		}

		shutdownCtx, shutdownCancel := context.WithTimeout(gCtx, 10*time.Second)
		defer shutdownCancel()

		return srv.Shutdown(shutdownCtx)
	})

	g.Go(func() error {
		return srv.ListenAndServe()
	})

	g.Go(func() error {
		if err := ht.Trip(ctx); err != nil {
			log.Printf("unexpected trip error: %v\n", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Printf("unexpected wait error: %v\n", err)
	}
}

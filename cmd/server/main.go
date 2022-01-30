package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"

	servertransporthttp "github.com/andream16/circuit-breaker-example/internal/server/transport/http"
	transporthttp "github.com/andream16/circuit-breaker-example/internal/transport/http"
)

type (
	httpServer interface {
		ListenAndServe() error
		Shutdown(ctx context.Context) error
	}
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	var srv httpServer = transporthttp.NewServer(
		":8080",
		&servertransporthttp.CloseCircuitHandler{},
		&servertransporthttp.OpenCircuitHandler{},
		&servertransporthttp.TripHandler{},
		&servertransporthttp.MetricsHandler{},
	)

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

	if err := g.Wait(); err != nil {
		log.Printf("unexpected wait error: %v\n", err)
	}
}

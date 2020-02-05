// Package regserver is a wrapper around the handlers package,
// turning it into a proper http server
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	logger "github.com/ipfs/go-log"
	"github.com/qri-io/qri/registry/regserver/handlers"
)

var (
	log      = logger.Logger("regserver")
	adminKey string
)

func main() {

	logger.SetLogLevel("regserver", "info")
	ctx := context.Background()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Info("creating temporary registry")
	inst, reg, cleanup, err := NewTempRepoRegistry(ctx)
	if err != nil {
		log.Fatalf("creating temp registry: %s", err)
	}

	addBasicDataset(inst)

	s := http.Server{
		Addr:    ":" + port,
		Handler: handlers.NewRoutes(reg),
	}

	log.Infof("serving on: %s", s.Addr)
	go func() {
		if err := s.ListenAndServe(); err != nil {
			log.Info(err.Error())
		}
	}()

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Info("Received signal, shutting down...")
	s.Close()
	cleanup()
	log.Info("done")
}

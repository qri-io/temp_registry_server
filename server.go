// Package regserver is a wrapper around the handlers package,
// turning it into a proper http server
package main

import (
	"context"
	"flag"
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

	logLevel  string
	port      string
	noCleanup bool
)

func init() {
	flag.StringVar(&port, "port", "2500", "port to listen on. default: `2500`")
	flag.BoolVar(&noCleanup, "no-cleanup", false, "don't remove directory on close")
	flag.StringVar(&logLevel, "log-level", "info", "set the remote's log level ['info', 'debug', 'warn', 'error']")
}

func main() {
	flag.Parse()

	logger.SetLogLevel("regserver", "info")
	ctx := context.Background()

	log.Info("creating temporary registry")
	inst, reg, cleanup, err := NewTempRepoRegistry(ctx, logLevel)
	if err != nil {
		log.Fatalf("creating temp registry: %s", err)
	}

	if err := createSynthsDataset(ctx, inst); err != nil {
		log.Fatal(err)
	}

	mux := handlers.NewRoutes(reg)
	mux.Handle("/sim/action", SimActionHandler(inst))

	s := http.Server{
		Addr:    ":" + port,
		Handler: mux,
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
	if !noCleanup {
		log.Infof("removing registry data")
		cleanup()
	}
	log.Info("done")
}

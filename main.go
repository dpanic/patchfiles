package main

import (
	"context"
	"embed"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"patchfiles/generator"
	"patchfiles/logger"
	"patchfiles/parser"

	"go.uber.org/zap"
)

var (
	//go:embed patches/*.yaml
	content embed.FS
	verbose = flag.Bool("verbose", true, "disable or enable verbose")
)

const (
	contextTimeout = 10 * time.Second
)

func main() {
	log, _ := logger.Setup(*verbose)
	defer log.Sync()
	log.Info("patchfiles started")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	// determine environment
	environment := os.Getenv("environment")
	environment = strings.ToLower(environment)
	environment = strings.Trim(environment, " ")
	if environment == "" {
		environment = "dev"
	}
	generator.Open(log, environment)

	// setup context timeout
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, contextTimeout)
	defer cancel()

	// gracefoul shutdown
	go func() {
		s := <-signals
		log.Warn("received signal",
			zap.String("signal", s.String()),
		)
		cancel()
		os.Exit(1)
	}()

	// run parser
	errors, results := parser.Run(log, &cancel, content)

	for {
		select {
		case e := <-errors:
			log = log.WithOptions(zap.Fields(
				zap.Error(e.Error),
				zap.String("fileLoc", *e.FileLoc),
			))
			log.Error("received error")

		case r := <-results:
			log = log.WithOptions(zap.Fields(
				zap.String("fileLoc", *r.FileLoc),
				zap.String("name", r.Name),
			))
			log.Info("received result")
			generator.Write(r, environment, log)

		case <-ctx.Done():
			log.Info("context is done")
			generator.Close()
			os.Exit(0)
		}
	}
}

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

	stats := map[string]int{
		"errors": 0,
		"good":   0,
		"total":  0,
	}

	for {
		select {
		case e := <-errors:
			logger := log.WithOptions(zap.Fields(
				zap.Error(e.Error),
				zap.String("fileLoc", *e.FileLoc),
			))
			logger.Error("received error")
			stats["errors"] += 1
			stats["total"] += 1

		case r := <-results:
			logger := log.WithOptions(zap.Fields(
				zap.String("fileLoc", *r.FileLoc),
				zap.String("name", r.Name),
			))
			logger.Info("received result")

			generator.Write(r, environment, log)
			stats["good"] += 1
			stats["total"] += 1

		case <-ctx.Done():
			log.Info("context is done")

			log.Debug("processing is done. stats",
				zap.Int("total", stats["total"]),
				zap.Int("good", stats["good"]),
				zap.Int("errors", stats["errors"]),
			)

			generator.Close()
			os.Exit(0)
		}
	}
}

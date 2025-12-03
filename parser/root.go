// Package parser parses YAML patch definitions and provides parsing results.
package parser

import (
	"context"
	"embed"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	patchesDir = "patches"
)

// Error represents a parsing error with the file location where it occurred.
type Error struct {
	Error   error   // The parsing error that occurred
	FileLoc *string // Location of the file that caused the error
}

// Result represents a successfully parsed patch file with its metadata.
type Result struct {
	Name    string  // Name of the patch (derived from filename)
	FileLoc *string // Location of the parsed YAML file
	Patch   *Patch  // Parsed patch definition
}

func Run(log *zap.Logger, cancel *context.CancelFunc, content embed.FS) (errors chan *Error, results chan *Result) {
	var res []fs.DirEntry
	errors = make(chan *Error, 100)
	results = make(chan *Result, 100)

	res, err := content.ReadDir(patchesDir)
	if err != nil {
		e := Error{
			Error:   err,
			FileLoc: nil,
		}
		errors <- &e
		return
	}

	go func() {
		for _, entry := range res {
			fileLoc := filepath.Join(patchesDir, entry.Name())
			fileName := filepath.Base(entry.Name())
			fileName = strings.Split(fileName, ".")[0]

			logger := log.WithOptions(zap.Fields(
				zap.String("fileLoc", fileLoc),
			))
			logger.Debug("attempt to parse file")

			body, err := content.ReadFile(fileLoc)
			if err != nil {
				logger.Error("error in reading",
					zap.Error(err),
				)

				e := Error{
					Error:   err,
					FileLoc: &fileLoc,
				}
				errors <- &e
				continue
			}

			patch, err := parse(body)
			if err != nil {
				logger.Error("error in parsing",
					zap.Error(err),
				)

				e := Error{
					Error:   err,
					FileLoc: &fileLoc,
				}
				errors <- &e
				continue
			}

			r := Result{
				Name:    fileName,
				FileLoc: &fileLoc,
				Patch:   patch,
			}

			logger.Info("successfully parsed file")
			results <- &r
		}

		// wait until all processed
		for len(results) > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		(*cancel)()
	}()

	return
}

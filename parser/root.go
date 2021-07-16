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

type Error struct {
	Error   error
	FileLoc *string
}

type Result struct {
	Name    string
	FileLoc *string
	Patch   *Patch
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

			log := log.WithOptions(zap.Fields(
				zap.String("fileLoc", fileLoc),
			))
			log.Debug("attempt to parse file")

			body, err := content.ReadFile(fileLoc)
			if err != nil {
				log.Error("error in reading",
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
				log.Error("error in parsing",
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

			log.Info("successfully parsed file")
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

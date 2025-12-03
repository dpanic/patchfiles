// Package generator generates patch and revert bash scripts from YAML patch definitions.
package generator

import (
	"fmt"
	"os"
	"strings"

	"patchfiles/parser"

	"go.uber.org/zap"
)

type Generator struct {
	Log         *zap.Logger
	Environment string

	n          map[string]string
	names      []string
	c          map[string]string
	categories []string
	fdPatch    *os.File
	fdRevert   *os.File
}

// Open opens both patch and revert file descriptors
func (generator *Generator) Open() {
	files := []string{
		"patch",
		"revert",
	}

	generator.n = make(map[string]string)
	generator.c = make(map[string]string)

	for _, name := range files {
		fileLoc := fmt.Sprintf("%s.sh", name)
		if generator.Environment == "dev" {
			fileLoc = fmt.Sprintf("%s_dev.sh", name)
		}
		fd, err := os.Create(fileLoc)
		os.Chmod(fileLoc, 0o755)

		if name == "patch" {
			generator.fdPatch = fd
		} else {
			generator.fdRevert = fd
		}

		if err != nil {
			generator.Log.Error("error in opening file",
				zap.Error(err),
				zap.String("fileLoc", fileLoc),
			)
		} else {
			action := fmt.Sprintf("%sING", strings.ToUpper(name))
			err = generator.writeHeader(fd, action)
			if err != nil {
				generator.Log.Error("error in writing header",
					zap.String("fileLoc", fileLoc),
					zap.Error(err),
				)
			}
		}
	}
}

// Close closes opened file descriptors for
func (generator *Generator) Close() {
	for name := range generator.n {
		generator.names = append(generator.names, name)
	}
	for category := range generator.c {
		generator.categories = append(generator.categories, category)
	}

	files := []string{
		"patch",
		"revert",
	}
	for _, name := range files {
		action := fmt.Sprintf("%sING", strings.ToUpper(name))
		var err error
		if name == "patch" {
			err = generator.writeFooter(generator.fdPatch, action)
		} else {
			err = generator.writeFooter(generator.fdRevert, action)
		}

		if err != nil {
			generator.Log.Error("error in writing footer",
				zap.Error(err),
			)
		}
	}

	if generator.fdPatch != nil {
		generator.fdPatch.Sync()
		generator.fdPatch.Close()
	}

	if generator.fdRevert != nil {
		generator.fdRevert.Sync()
		generator.fdRevert.Close()
	}
}

// Save generates output for a patch and revert
func (generator *Generator) Write(p *parser.Result) {
	generator.n[p.Name] = ""
	for _, category := range p.Patch.Categories {
		generator.c[category] = ""
	}

	err := generator.writePatch(p)
	if err != nil {
		generator.Log.Error("error in writing patch file",
			zap.Error(err),
		)
	}

	err = generator.writeRevert(p)
	if err != nil {
		generator.Log.Error("error in writing revert file",
			zap.Error(err),
		)
	}
}

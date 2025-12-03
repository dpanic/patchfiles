// Package generator generates patch and revert bash scripts from YAML patch definitions.
package generator

import (
	"fmt"
	"os"
	"strings"

	"patchfiles/parser"

	"go.uber.org/zap"
)

// Generator manages the generation of patch and revert bash scripts from YAML definitions.
type Generator struct {
	Log         *zap.Logger // Logger instance for logging operations
	Environment string      // Environment name (dev, prod, etc.)

	n          map[string]string // Map of patch names for tracking
	names      []string          // List of all patch names
	c          map[string]string // Map of categories for tracking
	categories []string          // List of all categories
	fdPatch    *os.File          // File descriptor for patch script
	fdRevert   *os.File          // File descriptor for revert script
}

// Open creates and opens file descriptors for both patch and revert bash scripts.
// It creates patch.sh (or patch_dev.sh in dev environment) and revert.sh (or revert_dev.sh).
// Each file is created with executable permissions and gets a header written to it.
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

// Close writes footers to both patch and revert scripts, then closes and syncs the file descriptors.
// It collects all patch names and categories for the footer help output before closing.
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

// Write generates output for a patch and revert script.
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

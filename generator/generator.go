package generator

import (
	"fmt"
	"os"
	"patchfiles/parser"
	"strings"

	"go.uber.org/zap"
)

var (
	fdPatch  *os.File
	fdRevert *os.File
)

// Open opens both patch and revert file descriptors
func Open(log *zap.Logger, environment string) {
	files := []string{
		"patch",
		"revert",
	}

	for _, name := range files {
		fileLoc := fmt.Sprintf("%s_%s.sh", name, environment)
		fd, err := os.Create(fileLoc)

		if name == "patch" {
			fdPatch = fd
		} else {
			fdRevert = fd
		}

		if err != nil {
			log.Error("error in opening file",
				zap.Error(err),
				zap.String("fileLoc", fileLoc),
			)
		} else {
			var head string

			action := fmt.Sprintf("%sING", strings.ToUpper(name))
			head, err = generateHeader(action, environment)
			if err != nil {
				log.Error("error in generating header",
					zap.String("fileLoc", fileLoc),
					zap.Error(err),
				)
			}

			fd.WriteString(head + "\n")
			fd.Sync()
		}
	}

}

// Close closes opened file descriptors for
func Close() {
	if fdPatch != nil {
		fdPatch.WriteString(fmt.Sprintf("echo 1 > %s\n", patchFilesControlFile))
		fdPatch.Sync()
		fdPatch.Close()
	}

	if fdRevert != nil {
		fdRevert.Sync()
		fdRevert.Close()
	}
}

// Save generates output for a patch and revert
func Write(p *parser.Result, environment string, log *zap.Logger) {
	err := writePatch(p, environment, log)
	if err != nil {
		log.Error("error in patching",
			zap.Error(err),
		)
	}

}

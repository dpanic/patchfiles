package generator

import (
	"encoding/base64"
	"fmt"
	"os"
	"patchfiles/parser"
	"strings"
	"time"

	"go.uber.org/zap"
)

var (
	fdPatch  *os.File
	fdRevert *os.File
)

const (
	header = `
	#!/bin/bash
	#
	# PATCHFILES
	# 
	# author: %s
	# version: %s
	# environment: %s
	# built: %s
	#
	#
	`
)

// generateHeader generates header based on input parameters
func generateHeader(environment string) string {
	built := time.Now().UTC().Format("2006-01-02 15:04:05 -07:00")

	author := os.Getenv("author")
	author = strings.ToLower(author)
	author = strings.Trim(author, " ")

	version := os.Getenv("version")
	version = strings.ToLower(version)
	version = strings.Trim(version, " ")

	head := strings.ReplaceAll(header, "\t", "")
	return fmt.Sprintf(head, author, version, environment, built)
}

func Open(log *zap.Logger, environment string) {
	var err error

	fileLoc := fmt.Sprintf("patch_%s.sh", environment)
	fdPatch, err = os.Create(fileLoc)
	if err != nil {
		log.Error("error in opening patch file",
			zap.Error(err),
			zap.String("fileLoc", fileLoc),
		)
	} else {
		fdPatch.WriteString(generateHeader(environment) + "\n")
		fdPatch.Sync()
	}

	fileLoc = fmt.Sprintf("revert_%s.sh", environment)
	fdRevert, err = os.Create(fileLoc)
	if err != nil {
		log.Error("error in opening revert file",
			zap.Error(err),
			zap.String("fileLoc", fileLoc),
		)
	} else {
		fdRevert.WriteString(generateHeader(environment) + "\n")
		fdRevert.Sync()
	}

}

// Close closes opened file descriptors for
func Close() {
	if fdPatch != nil {
		fdPatch.Close()
		fdPatch.Sync()
	}

	if fdRevert != nil {
		fdRevert.Sync()
		fdRevert.Close()
	}
}

// Save generates output for a patch and revert
func Write(p *parser.Result, environment string, log *zap.Logger) {
	log = log.WithOptions(zap.Fields(
		zap.String("fileLoc", *p.FileLoc),
		zap.String("name", p.Name),
	))
	log.Debug("attempt to write to file")

	// write patch
	bodyCommented := ""
	tmp := strings.Split(p.Patch.Body, "\n")
	for _, t := range tmp {
		bodyCommented += fmt.Sprintf("#    %s\n", t)
	}
	bodyCommented = strings.Trim(bodyCommented, "\n")

	p.Patch.Body = fmt.Sprintf("%s PATCHFILES START\n%s\n%s PATCHFILES END\n", p.Patch.CommentCharacter, p.Patch.Body, p.Patch.CommentCharacter)

	sEnc := base64.StdEncoding.EncodeToString([]byte(p.Patch.Body))
	writeMode := ">"
	if p.Patch.Mode == "append" {
		writeMode = ">>"
	}

	body := ""
	body += "#\n"
	body += fmt.Sprintf("# command '%s'\n#\n", p.Name)
	body += fmt.Sprintf("# description:\n#    %s\n#\n", p.Patch.Description)
	body += fmt.Sprintf("# body:\n%s\n", bodyCommented)
	body += strings.Repeat("#\n", 1)

	body += fmt.Sprintf("echo \"Patching '%s'\"\n", p.Name)
	body += fmt.Sprintf("echo \"%s\" | base64 -d - %s %s\n\n", sEnc, writeMode, p.Patch.Output)

	for _, command := range p.Patch.ExecuteAfter {
		body += fmt.Sprintf("%s\n", command)
	}

	fdPatch.WriteString(body + "\n")
	fdPatch.Sync()

	// write revert
}

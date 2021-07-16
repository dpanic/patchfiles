package generator

import (
	"bytes"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	templateHeader = `
	#!/bin/bash
	#
	# PATCHFILES SCRIPT FOR {{.ScriptFor}}
	# 
	# author: {{.Author}}
	# version: {{.Version}}
	# environment: {{.Environment}}
	# built: {{.Built}}
	#
	#
	`
)

type Header struct {
	ScriptFor   string
	Author      string
	Version     string
	Environment string
	Built       string
}

// generateHeader generates header based on input parameters
func generateHeader(scriptFor, environment string) (res string, err error) {
	built := time.Now().UTC().Format("2006-01-02 15:04:05 -07:00")

	author := os.Getenv("author")
	author = strings.ToLower(author)
	author = strings.Trim(author, " ")

	version := os.Getenv("version")
	version = strings.ToLower(version)
	version = strings.Trim(version, " ")

	data := Header{
		Author:      author,
		Version:     version,
		Built:       built,
		ScriptFor:   scriptFor,
		Environment: environment,
	}

	var (
		buf = new(bytes.Buffer)
	)

	tpl, err := template.New("template").Parse(templateHeader)

	t := template.Must(tpl, err)
	err = t.Execute(buf, data)
	if err != nil {
		return
	}

	res = buf.String()
	res = strings.ReplaceAll(res, "\t", "")
	return
}

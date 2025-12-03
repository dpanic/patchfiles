// Package parser parses YAML patch definitions and provides parsing results.
package parser

import (
	"gopkg.in/yaml.v2"
)

// Patch represents a patch definition parsed from a YAML file.
//
//go:generate easytags $GOFILE yaml:camel
type Patch struct {
	Output           string   `yaml:"output"`           // Target file path where patch will be applied
	Mode             string   `yaml:"mode"`             // Write mode: "overwrite" or "append"
	Body             string   `yaml:"body"`             // Content to write to the target file
	CommandsAfter    []string `yaml:"commandsAfter"`    // Commands to execute after applying the patch
	CommentCharacter string   `yaml:"commentCharacter"` // Character used for comments in target file
	Categories       []string `yaml:"categories"`       // List of categories this patch belongs to
	Description      string   `yaml:"description"`      // Human-readable description of the patch
}

// parse unmarshals YAML content into a Patch structure.
// It takes raw YAML bytes and returns a parsed Patch struct or an error if parsing fails.
func parse(body []byte) (patch *Patch, err error) {
	err = yaml.Unmarshal(body, &patch)

	return
}

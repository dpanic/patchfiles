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
	// CreatesFile marks a patch that introduces a brand new file rather than replacing an
	// existing one. Only those may be deleted on revert when no backup exists - doing it
	// unconditionally would erase files like sshd_config on a host that was never patched.
	CreatesFile bool `yaml:"createsFile"`
}

// parse unmarshals YAML content into a Patch structure.
// It takes raw YAML bytes and returns a parsed Patch struct or an error if parsing fails.
func parse(body []byte) (patch *Patch, err error) {
	err = yaml.Unmarshal(body, &patch)

	return
}

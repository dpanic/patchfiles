package parser

import (
	"gopkg.in/yaml.v2"
)

//go:generate easytags $GOFILE yaml:camel
type Patch struct {
	Output           string   `yaml:"output"`
	Mode             string   `yaml:"mode"`
	Body             string   `yaml:"body"`
	CommandsAfter    []string `yaml:"commandsAfter"`
	CommentCharacter string   `yaml:"commentCharacter"`
	Description      string   `yaml:"description"`
}

// parse function parses desired file to structure
func parse(body []byte) (patch *Patch, err error) {
	err = yaml.Unmarshal(body, &patch)

	return
}

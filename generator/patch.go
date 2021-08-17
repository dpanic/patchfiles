package generator

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"patchfiles/parser"
	"strings"
	"text/template"

	"go.uber.org/zap"
)

type PatchItem struct {
	NameShort        string
	NameLong         string
	Description      string
	Body             string
	Payload          string
	WriteMode        string
	Output           string
	Categories       []string
	CategoriesIfCase string
	CommandsAfter    []string
}

const (
	patchFilesControlFile = "/patchfile"
	templatePatchItem     = `
	#
	# COMMAND '{{.NameLong}}'
	# 
	# Categories: '{{.Categories}}'
	#
	# description:
	#    {{.Description}}
	#
	# body:
	{{.Body}}
	#
	
	if [[ "$category" == "all" || "$category" == "{{.NameShort}}" {{.CategoriesIfCase}} ]]; then
		echo -e "\n\n\n";
		echo "Patching '{{.NameLong}}'";
		
		echo "{{.Payload}}" | base64 -d - {{.WriteMode}} {{.Output}}

		{{ range $command := .CommandsAfter }}
			{{$command}}
		{{ end }}
	fi
`
)

func writePatch(p *parser.Result, environment string, log *zap.Logger) (err error) {
	logger := log.WithOptions(zap.Fields(
		zap.String("fileLoc", *p.FileLoc),
		zap.String("name", p.Name),
	))
	logger.Debug("attempt to write patch")

	// generate body commented
	bodyCommented := ""
	tmp := strings.Split(p.Patch.Body, "\n")
	for _, t := range tmp {
		bodyCommented += fmt.Sprintf("#    %s\n", t)
	}
	bodyCommented = strings.Trim(bodyCommented, "\n")

	// generate payload
	if p.Patch.Mode == "append" {
		p.Patch.Body = fmt.Sprintf("\n%s PATCHFILES START\n%s\n%s PATCHFILES END\n", p.Patch.CommentCharacter, p.Patch.Body, p.Patch.CommentCharacter)
	}
	payload := base64.StdEncoding.EncodeToString([]byte(p.Patch.Body + "\n"))

	// write mode
	commandsAfter := p.Patch.CommandsAfter
	writeMode := ">"
	if p.Patch.Mode == "append" {
		writeMode = ">>"
	} else {
		command := fmt.Sprintf("cp -r %s %s.oldpatchfile", p.Patch.Output, p.Patch.Output)
		commandsAfter = append(commandsAfter, command)
	}

	// prepare categories if case
	categories := make([]string, 0)
	for _, category := range p.Patch.Categories {
		categories = append(categories, fmt.Sprintf("\"$category\" == \"%s\"", category))
	}
	categoriesIfCase := strings.Join(categories, " || ")
	categoriesIfCase = strings.Trim(categoriesIfCase, " ")
	if categoriesIfCase != "" {
		categoriesIfCase = " || " + categoriesIfCase
	}

	var (
		buf = new(bytes.Buffer)
	)
	tpl, err := template.New("template").Parse(templatePatchItem)
	if err != nil {
		return
	}

	nameShort := strings.Split(p.Name, "_")[0]

	data := PatchItem{
		NameLong:         p.Name,
		NameShort:        nameShort,
		Description:      p.Patch.Description,
		Body:             bodyCommented,
		WriteMode:        writeMode,
		Output:           p.Patch.Output,
		Payload:          payload,
		CommandsAfter:    commandsAfter,
		Categories:       p.Patch.Categories,
		CategoriesIfCase: categoriesIfCase,
	}

	t := template.Must(tpl, err)
	err = t.Execute(buf, data)
	if err != nil {
		return
	}

	body := buf.String()
	body = strings.ReplaceAll(body, "\t", "")
	fdPatch.WriteString(body + "\n")
	fdPatch.Sync()
	return
}

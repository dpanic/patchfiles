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
	Name          string
	Description   string
	Body          string
	Payload       string
	WriteMode     string
	Output        string
	CommandsAfter []string
}

const (
	patchFilesControlFile = "/patchfile"
	templatePatchItem     = `
	#
	# COMMAND '{{.Name}}'
	#
	# description:
	#    {{.Description}}
	#
	# body:
	{{.Body}}
	#
	
	echo -e "\n\n\n"
	echo "Patching '{{.Name}}'"
	echo "{{.Payload}}" | base64 -d - {{.WriteMode}} {{.Output}}

	{{ range $command := .CommandsAfter }}
		{{$command}}
	{{ end }}
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

	var (
		buf = new(bytes.Buffer)
	)
	tpl, err := template.New("template").Parse(templatePatchItem)
	if err != nil {
		return
	}

	data := PatchItem{
		Name:          p.Name,
		Description:   p.Patch.Description,
		Body:          bodyCommented,
		WriteMode:     writeMode,
		Output:        p.Patch.Output,
		Payload:       payload,
		CommandsAfter: commandsAfter,
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

package generator

import (
	"bytes"
	"fmt"
	"patchfiles/parser"
	"strings"
	"text/template"

	"go.uber.org/zap"
)

type RevertItem struct {
	Name             string
	Description      string
	Categories       []string
	CategoriesIfCase string
	Command          string
	CommandsAfter    []string
}

const (
	templateRevertItem = `
	#
	# COMMAND '{{.Name}}'
	#
	# Categories: '{{.Categories}}'
	#
	#
	# description:
	#    {{.Description}}
	#


	if [[ "$category" == "all" {{.CategoriesIfCase}} ]]; then
		echo -e "\n\n\n"
		echo "Reverting '{{.Name}}'"

		{{.Command}}
		{{ range $command := .CommandsAfter }}
			{{$command}}
		{{ end }}
	fi;
`
)

func writeRevert(p *parser.Result, environment string, log *zap.Logger) (err error) {
	logger := log.WithOptions(zap.Fields(
		zap.String("fileLoc", *p.FileLoc),
		zap.String("name", p.Name),
	))
	logger.Debug("attempt to write revert")

	// generate payload
	start := fmt.Sprintf("%s PATCHFILES START", p.Patch.CommentCharacter)
	end := fmt.Sprintf("%s PATCHFILES END", p.Patch.CommentCharacter)

	// write mode
	writeMode := ">"
	if p.Patch.Mode == "append" {
		writeMode = ">>"
	}

	command := ""
	if writeMode == ">" {
		command = fmt.Sprintf("mv %s.oldpatchfile %s", p.Patch.Output, p.Patch.Output)
	} else {
		command += fmt.Sprintf("sed -i -e '/%s/,/%s/c\\' %s", start, end, p.Patch.Output)
	}

	var (
		buf = new(bytes.Buffer)
	)
	tpl, err := template.New("template").Parse(templateRevertItem)
	if err != nil {
		return
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

	data := RevertItem{
		Name:             p.Name,
		Description:      p.Patch.Description,
		Command:          command,
		CommandsAfter:    p.Patch.CommandsAfter,
		CategoriesIfCase: categoriesIfCase,
	}

	t := template.Must(tpl, err)
	err = t.Execute(buf, data)
	if err != nil {
		return
	}

	body := buf.String()
	body = strings.ReplaceAll(body, "\t", "")
	fdRevert.WriteString(body + "\n")
	fdRevert.Sync()

	return
}

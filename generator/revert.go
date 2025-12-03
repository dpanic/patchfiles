package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"patchfiles/parser"

	"go.uber.org/zap"
)

// RevertItem contains template data for generating a single revert command in the bash script.
type RevertItem struct {
	NameShort        string   // Short name of the patch (first part before underscore)
	NameLong         string   // Full name of the patch
	Description      string   // Human-readable description of the patch
	Categories       []string // List of categories this patch belongs to
	CategoriesIfCase string   // Generated if-case string for category matching
	Command          string   // Bash command to revert the patch
	CommandsAfter    []string // Commands to execute after reverting the patch
}

const (
	templateRevertItem = `
	#
	# COMMAND '{{.NameLong}}'
	#
	# Categories: '{{.Categories}}'
	#
	#
	# description:
	#    {{.Description}}
	#


	if [[ "$category" == "all" || "$category" == "{{.NameShort}}" {{.CategoriesIfCase}} ]]; then
		echo -e "\n\n\n"
		echo "Reverting '{{.NameLong}}'"

		{{.Command}}
		{{ range $command := .CommandsAfter }}
			{{$command}}
		{{ end }}
	fi;
`
)

func (generator *Generator) writeRevert(p *parser.Result) (err error) {
	logger := generator.Log.WithOptions(zap.Fields(
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

	buf := new(bytes.Buffer)
	tpl, err := template.New("template").Parse(templateRevertItem)
	if err != nil {
		return
	}

	nameShort := strings.Split(p.Name, "_")[0]

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
		NameLong:         p.Name,
		NameShort:        nameShort,
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
	generator.fdRevert.WriteString(body + "\n")
	generator.fdRevert.Sync()

	return
}

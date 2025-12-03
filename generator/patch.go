package generator

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"patchfiles/parser"

	"go.uber.org/zap"
)

// PatchItem contains template data for generating a single patch command in the bash script.
type PatchItem struct {
	NameShort        string   // Short name of the patch (first part before underscore)
	NameLong         string   // Full name of the patch
	Description      string   // Human-readable description of the patch
	Body             string   // Commented body content for display in generated script
	Payload          string   // Base64-encoded payload to write to target file
	WriteMode        string   // Bash write mode: ">" for overwrite, ">>" for append
	Output           string   // Target file path where patch will be applied
	Categories       []string // List of categories this patch belongs to
	CategoriesIfCase string   // Generated if-case string for category matching
	CommandsAfter    []string // Commands to execute after applying the patch
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
		
		SKIP_PATCH=0
		{{ if eq .WriteMode ">>" }}
		# Check if already patched (append mode)
		if grep -q "PATCHFILES START" "{{.Output}}" 2>/dev/null; then
			echo "Warning: '{{.NameLong}}' appears to be already patched. Skipping to avoid duplicates."
			echo "If you want to re-apply, use revert first or manually remove PATCHFILES START/END blocks."
			SKIP_PATCH=1
		fi
		{{ else }}
		# Check if already patched (overwrite mode)
		if [ -f "{{.Output}}.oldpatchfile" ]; then
			echo "Warning: '{{.NameLong}}' appears to be already patched (backup file exists). Skipping to avoid overwriting backup."
			echo "If you want to re-apply, use revert first or manually remove {{.Output}}.oldpatchfile"
			SKIP_PATCH=1
		fi
		{{ end }}
		
		if [ "$SKIP_PATCH" -eq 0 ]; then
			echo "{{.Payload}}" | base64 -d - {{.WriteMode}} {{.Output}}

			{{ range $command := .CommandsAfter }}
				{{$command}}
			{{ end }}
		fi
		
		echo "{{.Payload}}" | base64 -d - {{.WriteMode}} {{.Output}}

		{{ range $command := .CommandsAfter }}
			{{$command}}
		{{ end }}
	fi
`
)

func (generator *Generator) writePatch(p *parser.Result) (err error) {
	logger := generator.Log.WithOptions(zap.Fields(
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

	buf := new(bytes.Buffer)
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
	generator.fdPatch.WriteString(body + "\n")
	generator.fdPatch.Sync()
	return
}

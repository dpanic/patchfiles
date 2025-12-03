package generator

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"go.uber.org/zap"
)

// Footer contains template data for generating script footers.
type Footer struct {
	Names                 []string // List of all patch names for help output
	Categories            []string // List of all categories for help output
	ScriptFor             string   // Action type: "PATCHING" or "REVERTING"
	PatchFilesControlFile string   // Path to control file that tracks patch status
}

const (
	// templateFooter is the bash script template for the footer section of patch/revert scripts.
	templateFooter = `
	function help_me() {
		echo -e "\n\n";
		echo "******************";
		echo "*** How to use ***";
		echo "******************";

		echo -e "\n";
		echo "Available categories are:";
		{{ range $category := .Categories }}
			echo "* {{$category}}";
		{{ end }}
		echo -e "\n";

		echo "Available patches are:";
		{{ range $name := .Names }}
			echo "* {{$name}}";
		{{ end }}

		echo -e "\n";
		echo "Examples:";
		echo "./patch.sh all";
		echo "./patch.sh security";
		echo "./patch.sh sshd";
		echo "./revert.sh sshd";
	}

	if [[ "$category" == "" || "$category" == "help" ]]; then
		help_me;
		exit 1;
	fi;

	{{ if eq .ScriptFor "PATCHING" }}
		echo 1 > {{.PatchFilesControlFile}};
	{{ end }}	

	{{ if eq .ScriptFor "REVERTING" }}
		rm -f {{.PatchFilesControlFile}};
	{{ end }}	
`
)

// writeFooter generates and writes the bash script footer to the given file descriptor.
// It includes a help function, category/patch listing, and logic to create/remove the control file
// that tracks whether the system has been patched.
func (generator *Generator) writeFooter(fd *os.File, scriptFor string) (err error) {
	logger := generator.Log.WithOptions(zap.Fields())
	logger.Debug("attempt to write footer",
		zap.String("scriptFor", scriptFor),
	)

	buf := new(bytes.Buffer)

	tpl, err := template.New("template").Parse(templateFooter)

	obj := Footer{
		ScriptFor:             scriptFor,
		Names:                 generator.names,
		Categories:            generator.categories,
		PatchFilesControlFile: patchFilesControlFile,
	}

	t := template.Must(tpl, err)
	err = t.Execute(buf, obj)
	if err != nil {
		return
	}

	res := buf.String()
	res = strings.ReplaceAll(res, "\t", "")

	fd.WriteString(res + "\n")
	fd.Sync()

	return
}

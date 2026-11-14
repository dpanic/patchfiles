package generator

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"patchfiles/parser"

	"go.uber.org/zap"
)

// runBash runs script with args and returns its combined output.
func runBash(t *testing.T, script string, args ...string) string {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	loc := filepath.Join(t.TempDir(), "script.sh")
	if err := os.WriteFile(loc, []byte(script), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("bash", append([]string{loc}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("bash %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// renderHeader renders the header template with a control file the test owns.
func renderHeader(t *testing.T, scriptFor, controlFile string) string {
	t.Helper()
	tpl := template.Must(template.New("header").Parse(templateHeader))
	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, Header{ScriptFor: scriptFor, PatchFilesControlFile: controlFile}); err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(buf.String(), "\t", "")
}

func TestHeaderGuardAppliesToAllOnly(t *testing.T) {
	tests := []struct {
		name      string
		scriptFor string
		marker    bool // control file exists before the run
		category  string
		wantRun   bool
	}{
		{"patch all on patched system exits", "PATCHING", true, "all", false},
		{"patch category on patched system runs", "PATCHING", true, "rps", true},
		{"patch all on fresh system runs", "PATCHING", false, "all", true},
		{"revert all on unpatched system exits", "REVERTING", false, "all", false},
		{"revert category on unpatched system runs", "REVERTING", false, "rps", true},
		{"revert all on patched system runs", "REVERTING", true, "all", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			control := filepath.Join(t.TempDir(), "patchfile")
			if tt.marker {
				if err := os.WriteFile(control, []byte("1\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			out := runBash(t, renderHeader(t, tt.scriptFor, control)+"\necho BODY_RAN\n", tt.category)
			if got := strings.Contains(out, "BODY_RAN"); got != tt.wantRun {
				t.Fatalf("body ran = %v, want %v; output:\n%s", got, tt.wantRun, out)
			}
		})
	}
}

// writePatchScript renders one patch block into a temp file and returns its text.
func writePatchScript(t *testing.T, p *parser.Result) string {
	t.Helper()
	fd, err := os.CreateTemp(t.TempDir(), "patch-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	g := &Generator{Log: zap.NewNop(), fdPatch: fd}
	if err := g.writePatch(p); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(fd.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func newResult(name, output, mode string) *parser.Result {
	loc := "patches/" + name + ".yaml"
	return &parser.Result{
		Name:    name,
		FileLoc: &loc,
		Patch: &parser.Patch{
			Output:           output,
			Mode:             mode,
			Body:             "key value",
			CommentCharacter: "#",
		},
	}
}

func TestPatchCreatesParentDirBeforeWrite(t *testing.T) {
	for _, mode := range []string{"overwrite", "append"} {
		t.Run(mode, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "missing", "dir", "target.conf")
			script := writePatchScript(t, newResult("demo_item", out, mode))

			mkdir := strings.Index(script, `mkdir -p "`+filepath.Dir(out)+`"`)
			write := strings.Index(script, "| base64 -d -")
			if mkdir < 0 || write < 0 || mkdir > write {
				t.Fatalf("mkdir at %d, write at %d; want mkdir before write:\n%s", mkdir, write, script)
			}
			if backup := strings.Index(script, "cp -a"); backup >= 0 && mkdir > backup {
				t.Fatalf("mkdir at %d after backup at %d", mkdir, backup)
			}

			runBash(t, "category=\"$1\"\n"+script, "demo")
			got, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("patch output not written: %v", err)
			}
			if !strings.Contains(string(got), "key value") {
				t.Fatalf("unexpected output content: %q", got)
			}
		})
	}
}

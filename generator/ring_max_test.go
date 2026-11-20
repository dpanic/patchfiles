package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"patchfiles/parser"

	"gopkg.in/yaml.v2"
)

// fakeEthtool keeps "rxmax rxcur txmax txcur" per NIC in ring/<nic> and logs every -G call.
const fakeEthtool = `#!/bin/sh
d=$(dirname "$0"); op=$1; nic=$2; shift 2
read -r rm rc tm tc < "$d/ring/$nic" || exit 71
case $op in
-g) printf 'Pre-set maximums:\nRX:\t\t%s\nRX Mini:\tn/a\nTX:\t\t%s\n' "$rm" "$tm"
    printf 'Current hardware settings:\nRX:\t\t%s\nRX Mini:\tn/a\nTX:\t\t%s\n' "$rc" "$tc" ;;
-G) echo "$nic $*" >> "$d/calls" ;;
esac
`

// loadPatch reads one patch definition from the patches directory.
func loadPatch(t *testing.T, name string) *parser.Patch {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "patches", name+".yaml"))
	if err != nil {
		t.Fatal(err)
	}
	p := new(parser.Patch)
	if err := yaml.Unmarshal(b, p); err != nil {
		t.Fatal(err)
	}
	return p
}

// A ring change resets the link of a running NIC, so ring-max may only touch a NIC that is down.
func TestRingMaxRaisesOnlyDownNICs(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	dir := t.TempDir()
	write := func(rel, body string, mode os.FileMode) {
		t.Helper()
		loc := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(loc), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(loc, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
	}
	write("ring-max", loadPatch(t, "ring_max").Body+"\n", 0o755)
	write("ethtool", fakeEthtool, 0o755)
	write("calls", "", 0o644)
	nics := map[string][2]string{ // interface flags (bit 0 = up), rings
		"down0": {"0x1002", "4096 256 4096 256"},
		"up0":   {"0x1003", "4096 256 4096 256"},
		"full0": {"0x1002", "4096 4096 4096 4096"},
	}
	for nic, st := range nics {
		write("net/"+nic+"/flags", st[0]+"\n", 0o644)
		write("net/"+nic+"/device", "", 0o644)
		write("ring/"+nic, st[1]+"\n", 0o644)
	}

	cmd := exec.Command("sh", filepath.Join(dir, "ring-max"))
	cmd.Env = append(os.Environ(),
		"RING_SYSNET="+filepath.Join(dir, "net"),
		"RING_ETHTOOL="+filepath.Join(dir, "ethtool"),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ring-max: %v\n%s", err, out)
	}
	calls, err := os.ReadFile(filepath.Join(dir, "calls"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(string(calls)), "down0 rx 4096 tx 4096"; got != want {
		t.Fatalf("ethtool -G calls = %q, want %q", got, want)
	}
	if !strings.Contains(string(out), "up0 is up, left alone") {
		t.Fatalf("running NIC not reported:\n%s", out)
	}
}

// autotune runs once at boot and saw only NICs that were already up; ring_max owns the rings.
func TestAutotuneLeavesRingsToRingMax(t *testing.T) {
	if strings.Contains(loadPatch(t, "autotune").Body, "ethtool") {
		t.Fatal("autotune must not touch NIC rings: ring_max sets them from udev, before the NIC is up")
	}
}

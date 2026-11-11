package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sohaha/zlsgo/zdi"
)

func TestParsePort(t *testing.T) {
	cases := map[string]int{
		"3788":           3788,
		":8080":          8080,
		"127.0.0.1:9090": 9090,
		"[::1]:9000":     9000,
		" 8080 ":         8080,
		"":               0,
		"not-a-port":     0,
	}

	for input, want := range cases {
		if got := parsePort(input); got != want {
			t.Fatalf("parsePort(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestNewConfReadsBaseConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "appconf")
	content := `[base]
port = "127.0.0.1:9090"
hot_reload = false
log_show_date = true
zone = 8
`
	if err := os.WriteFile(path+".toml", []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	oldName, oldBase, oldDefaults := ConfFileName, baseConf, DefaultConf
	ConfFileName = path
	baseConf = BaseConf{Zone: 8, Port: "3788", HotReload: true}
	DefaultConf = nil
	t.Cleanup(func() {
		ConfFileName, baseConf, DefaultConf = oldName, oldBase, oldDefaults
	})

	c := NewConf()(zdi.New())

	if c.Base.Port != "127.0.0.1:9090" {
		t.Fatalf("port = %q, want file value", c.Base.Port)
	}
	if c.Base.HotReload {
		t.Fatal("hot_reload from file should override the code default")
	}
	if !c.Base.LogShowDate {
		t.Fatal("log_show_date from file should be applied")
	}
}

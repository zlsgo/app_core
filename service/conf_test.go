package service

import (
	"os"
	"path/filepath"
	"strings"
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

type lateOptions struct {
	Prefix string `json:"prefix"`
	Title  string `json:"title"`
	Secret string `json:"-"`
}

func (lateOptions) ConfKey() string { return "late" }

func (lateOptions) DisableWrite() bool { return true }

func TestApplyDefaultsAfterConfBuilt(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lateconf")
	content := `[late]
prefix = "/from-file"
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

	opts := &lateOptions{Prefix: "/code", Title: "code-title", Secret: "in-memory"}
	DefaultConf = append(DefaultConf, opts)

	c.ApplyDefaults()

	if got := c.Get("late.prefix").String(); got != "/from-file" {
		t.Fatalf("late.prefix = %q, want file value", got)
	}
	if got := c.Get("late.title").String(); got != "code-title" {
		t.Fatalf("late.title = %q, want code default", got)
	}

	if opts.Prefix != "/from-file" {
		t.Fatalf("opts.Prefix = %q, want file value applied to struct", opts.Prefix)
	}
	if opts.Title != "code-title" {
		t.Fatalf("opts.Title = %q, want code default kept", opts.Title)
	}
	if opts.Secret != "in-memory" {
		t.Fatalf("opts.Secret = %q, want untouched", opts.Secret)
	}

	c.ApplyDefaults()
	if got := c.Get("late.prefix").String(); got != "/from-file" {
		t.Fatalf("after second ApplyDefaults late.prefix = %q", got)
	}
}

func TestApplyDefaultsValueType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lateconf2")
	if err := os.WriteFile(path+".toml", []byte(""), 0o644); err != nil {
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

	DefaultConf = append(DefaultConf, lateValueOptions{Driver: "sqlite"})
	c.ApplyDefaults()

	var loaded lateValueOptions
	if err := c.Unmarshal("latevalue", &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.Driver != "sqlite" {
		t.Fatalf("Driver = %q, want sqlite", loaded.Driver)
	}
}

type lateValueOptions struct {
	Driver string `json:"driver"`
}

func (lateValueOptions) ConfKey() string { return "latevalue" }

func (lateValueOptions) DisableWrite() bool { return true }

type writeOptions struct {
	Foo string `json:"foo"`
	Bar string `json:"bar"`
}

func (writeOptions) ConfKey() string { return "nowrite" }

func (writeOptions) DisableWrite() bool { return true }

type plainOptions struct {
	Name string `json:"name"`
}

func (plainOptions) ConfKey() string { return "plain" }

func (plainOptions) DisableWrite() bool { return false }

// withConfFile prepares a temp config file and points the global conf state at it.
func withConfFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "appconf")
	if content != "" {
		if err := os.WriteFile(path+".toml", []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldName, oldBase, oldDefaults := ConfFileName, baseConf, DefaultConf
	ConfFileName = path
	baseConf = BaseConf{Zone: 8, Port: "3788", HotReload: true}
	DefaultConf = nil
	t.Cleanup(func() {
		ConfFileName, baseConf, DefaultConf = oldName, oldBase, oldDefaults
	})

	return path
}

func readConfFile(t *testing.T, path string) string {
	t.Helper()

	b, err := os.ReadFile(path + ".toml")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestWriteSkipsDisableWriteDefaults(t *testing.T) {
	path := withConfFile(t, "")
	DefaultConf = []interface{}{
		&writeOptions{Foo: "code-foo", Bar: "code-bar"},
		plainOptions{Name: "code-name"},
	}

	c := NewConf()(zdi.New())
	if err := c.Write(); err != nil {
		t.Fatal(err)
	}

	content := readConfFile(t, path)
	if strings.Contains(content, "nowrite") || strings.Contains(content, "code-foo") {
		t.Fatalf("no-write config leaked into file:\n%s", content)
	}
	if !strings.Contains(content, "[plain]") || !strings.Contains(content, "code-name") {
		t.Fatalf("writable config missing from file:\n%s", content)
	}
	if got := c.Get("nowrite.foo").String(); got != "code-foo" {
		t.Fatalf("nowrite.foo = %q, want code default kept in memory", got)
	}
}

func TestWriteKeepsDisableWriteFromFile(t *testing.T) {
	path := withConfFile(t, `[nowrite]
foo = 'user-foo'

[base]
port = '127.0.0.1:9999'
zone = 8
`)
	DefaultConf = []interface{}{
		&writeOptions{Foo: "code-foo", Bar: "code-bar"},
		baseConf,
	}

	c := NewConf()(zdi.New())
	if err := c.Write(); err != nil {
		t.Fatal(err)
	}

	content := readConfFile(t, path)
	if !strings.Contains(content, "user-foo") {
		t.Fatalf("user value was hidden:\n%s", content)
	}
	if strings.Contains(content, "code-bar") || strings.Contains(content, "bar =") {
		t.Fatalf("no-write default leaked into file:\n%s", content)
	}
	if !strings.Contains(content, "127.0.0.1:9999") {
		t.Fatalf("base config was hidden:\n%s", content)
	}
}

func TestWritePersistsExplicitlySetNoWriteValue(t *testing.T) {
	path := withConfFile(t, "")
	DefaultConf = []interface{}{
		&writeOptions{Foo: "code-foo", Bar: "code-bar"},
		plainOptions{Name: "code-name"},
	}

	c := NewConf()(zdi.New())
	c.Set("nowrite.bar", "generated")
	if err := c.Write(); err != nil {
		t.Fatal(err)
	}

	content := readConfFile(t, path)
	if !strings.Contains(content, "generated") {
		t.Fatalf("explicitly set value was not persisted:\n%s", content)
	}
	if strings.Contains(content, "code-foo") {
		t.Fatalf("unset no-write default leaked into file:\n%s", content)
	}

	// Once written, the value comes from the file and stays on later writes.
	if err := c.Write(); err != nil {
		t.Fatal(err)
	}
	if content := readConfFile(t, path); !strings.Contains(content, "generated") {
		t.Fatalf("value lost on second write:\n%s", content)
	}
}

func TestWriteKeepsOnlyFileFieldsOfDisableWriteGroup(t *testing.T) {
	path := withConfFile(t, `[nowrite]
foo = 'user-foo'
`)
	DefaultConf = []interface{}{&writeOptions{Foo: "code-foo", Bar: "code-bar"}}

	c := NewConf()(zdi.New())
	if err := c.Write(); err != nil {
		t.Fatal(err)
	}

	content := readConfFile(t, path)
	if !strings.Contains(content, "user-foo") {
		t.Fatalf("user value was hidden:\n%s", content)
	}
	if strings.Contains(content, "bar") {
		t.Fatalf("no-write default field leaked into file:\n%s", content)
	}
}

func TestWritePrefersFileValueOfDisableWriteGroup(t *testing.T) {
	path := withConfFile(t, `[nowrite]
foo = 'startup-foo'
`)
	DefaultConf = []interface{}{&writeOptions{Foo: "code-foo", Bar: "code-bar"}}

	c := NewConf()(zdi.New())

	if err := os.WriteFile(path+".toml", []byte("[nowrite]\nfoo = 'edited-foo'\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c.Set("nowrite.bar", "generated")

	if err := c.Write(); err != nil {
		t.Fatal(err)
	}

	content := readConfFile(t, path)
	if !strings.Contains(content, "edited-foo") {
		t.Fatalf("file value was clobbered:\n%s", content)
	}
	if !strings.Contains(content, "generated") {
		t.Fatalf("explicitly set value was not persisted:\n%s", content)
	}
	if strings.Contains(content, "code-foo") {
		t.Fatalf("no-write default leaked into file:\n%s", content)
	}
}

func TestWriteKeepsEmptyDisableWriteTable(t *testing.T) {
	path := withConfFile(t, "[nowrite]\n")
	DefaultConf = []interface{}{&writeOptions{Foo: "code-foo", Bar: "code-bar"}}

	c := NewConf()(zdi.New())
	if err := c.Write(); err != nil {
		t.Fatal(err)
	}

	content := readConfFile(t, path)
	if !strings.Contains(content, "[nowrite]") {
		t.Fatalf("empty user table was hidden:\n%s", content)
	}
	if strings.Contains(content, "code-foo") {
		t.Fatalf("no-write default leaked into file:\n%s", content)
	}
}

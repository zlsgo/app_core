package service

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zlog"
)

type lifecycleProbe struct {
	App
	name    string
	calls   *[]string
	loadErr error
}

func (m *lifecycleProbe) record(stage string) {
	*m.calls = append(*m.calls, m.name+":"+stage)
	if m.App.DI == nil || m.App.Conf == nil || m.App.Log == nil {
		*m.calls = append(*m.calls, m.name+":missing-app")
	}
}

func (m *lifecycleProbe) Name() string                  { return m.name }
func (m *lifecycleProbe) Tasks() []Task                 { return nil }
func (m *lifecycleProbe) Controller() []Controller      { return nil }
func (m *lifecycleProbe) Load(zdi.Invoker) (any, error) { m.record("load"); return nil, m.loadErr }
func (m *lifecycleProbe) Start(zdi.Invoker) error       { m.record("start"); return nil }
func (m *lifecycleProbe) Done(zdi.Invoker) error        { m.record("done"); return nil }

func TestInitModuleLifecycle(t *testing.T) {
	di := zdi.New()
	calls := make([]string, 0, 8)
	a := &lifecycleProbe{name: "a", calls: &calls}
	b := &lifecycleProbe{name: "b", calls: &calls}
	app := &App{DI: di, Conf: &Conf{}, Log: zlog.New("")}

	if err := InitModule([]Module{b, a}, app); err != nil {
		t.Fatalf("InitModule: %v", err)
	}

	want := []string{"a:load", "b:load", "a:start", "b:start", "a:done", "b:done"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("unexpected lifecycle calls: got %v, want %v", calls, want)
	}
	if a.Conf != app.Conf || b.Conf != app.Conf {
		t.Fatal("module Conf was not injected")
	}
	if a.DI == nil || b.DI == nil {
		t.Fatal("module DI was not injected")
	}
	if a.Log == nil || b.Log == nil {
		t.Fatal("module Log was not injected")
	}
}

func TestInitModuleWrapsLoadError(t *testing.T) {
	di := zdi.New()
	calls := make([]string, 0, 2)
	bad := &lifecycleProbe{name: "bad", calls: &calls, loadErr: errors.New("boom")}
	app := &App{DI: di, Conf: &Conf{}, Log: zlog.New("")}

	err := InitModule([]Module{bad}, app)
	if err == nil {
		t.Fatal("expected InitModule to fail")
	}
	msg := strings.Join(zerror.UnwrapErrors(err), " ")
	if !strings.Contains(msg, "bad failed to Load") || !strings.Contains(msg, "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitModuleSkipsNilModules(t *testing.T) {
	di := zdi.New()
	calls := make([]string, 0, 3)
	good := &lifecycleProbe{name: "good", calls: &calls}
	var typedNil *lifecycleProbe
	app := &App{DI: di, Conf: &Conf{}, Log: zlog.New("")}

	if err := InitModule([]Module{nil, typedNil, good}, app); err != nil {
		t.Fatalf("InitModule: %v", err)
	}

	want := []string{"good:load", "good:start", "good:done"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestIsNilModule(t *testing.T) {
	var typedNil *lifecycleProbe
	if !IsNilModule(nil) {
		t.Fatal("nil module should be reported as nil")
	}
	if !IsNilModule(typedNil) {
		t.Fatal("typed nil module should be reported as nil")
	}
	if IsNilModule(&lifecycleProbe{}) {
		t.Fatal("non-nil module should not be reported as nil")
	}
}

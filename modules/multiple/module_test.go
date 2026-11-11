package multiple

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

type probeModule struct {
	service.App
	gotDI   bool
	gotConf bool
	gotLog  bool
}

func (p *probeModule) Name() string                     { return "ProbeModule" }
func (p *probeModule) Tasks() []service.Task            { return nil }
func (p *probeModule) Controller() []service.Controller { return nil }
func (p *probeModule) Start(zdi.Invoker) error          { return nil }
func (p *probeModule) Done(zdi.Invoker) error           { return nil }
func (p *probeModule) Load(zdi.Invoker) (any, error) {
	p.gotDI = p.App.DI != nil
	p.gotConf = p.App.Conf != nil
	p.gotLog = p.App.Log != nil
	return nil, nil
}

func TestLoadInjectsChildModules(t *testing.T) {
	di := zdi.New()
	child := &probeModule{}
	bundle := New(map[string]service.Module{"probe": child})
	bundle.App = service.App{DI: di, Conf: new(service.Conf), Log: zlog.New("")}

	if _, err := bundle.Load(di); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !child.gotDI || !child.gotConf || !child.gotLog {
		t.Fatalf("child module not fully injected: DI=%v Conf=%v Log=%v",
			child.gotDI, child.gotConf, child.gotLog)
	}
}

type orderedProbe struct {
	service.App
	name  string
	calls *[]string
}

func (p *orderedProbe) Name() string                     { return p.name }
func (p *orderedProbe) Tasks() []service.Task            { return nil }
func (p *orderedProbe) Controller() []service.Controller { return nil }
func (p *orderedProbe) Load(zdi.Invoker) (any, error) {
	*p.calls = append(*p.calls, p.name+":load")
	return nil, nil
}
func (p *orderedProbe) Start(zdi.Invoker) error {
	*p.calls = append(*p.calls, p.name+":start")
	return nil
}
func (p *orderedProbe) Done(zdi.Invoker) error {
	*p.calls = append(*p.calls, p.name+":done")
	return nil
}

func resetModules(t *testing.T) {
	t.Helper()
	old := modules
	modules = zarray.NewHashMap[string, service.Module]()
	t.Cleanup(func() { modules = old })
}

func TestLifecycleRunsInSortedKeyOrder(t *testing.T) {
	resetModules(t)

	di := zdi.New()
	calls := make([]string, 0, 9)
	bundle := New(map[string]service.Module{
		"c": &orderedProbe{name: "C", calls: &calls},
		"a": &orderedProbe{name: "A", calls: &calls},
		"b": &orderedProbe{name: "B", calls: &calls},
	})
	bundle.App = service.App{DI: di, Conf: new(service.Conf), Log: zlog.New("")}

	if got := bundle.Name(); got != "Multiple [A, B, C]" {
		t.Fatalf("Name() = %q, want %q", got, "Multiple [A, B, C]")
	}

	if _, err := bundle.Load(di); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := bundle.Start(di); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := bundle.Done(di); err != nil {
		t.Fatalf("Done: %v", err)
	}

	want := []string{
		"A:load", "B:load", "C:load",
		"A:start", "B:start", "C:start",
		"A:done", "B:done", "C:done",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestNilChildrenAreIgnored(t *testing.T) {
	resetModules(t)

	di := zdi.New()
	calls := make([]string, 0, 3)
	var typedNil *orderedProbe
	bundle := New(map[string]service.Module{
		"a": &orderedProbe{name: "A", calls: &calls},
		"b": nil,
		"c": typedNil,
	})
	bundle.App = service.App{DI: di, Conf: new(service.Conf), Log: zlog.New("")}

	if got := bundle.Name(); got != "Multiple [A]" {
		t.Fatalf("Name() = %q, want %q", got, "Multiple [A]")
	}
	if tasks := bundle.Tasks(); len(tasks) != 0 {
		t.Fatalf("Tasks() = %v, want empty", tasks)
	}
	if controllers := bundle.Controller(); len(controllers) != 0 {
		t.Fatalf("Controller() = %v, want empty", controllers)
	}
	if _, err := bundle.Load(di); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := bundle.Start(di); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := bundle.Done(di); err != nil {
		t.Fatalf("Done: %v", err)
	}

	want := []string{"A:load", "A:start", "A:done"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

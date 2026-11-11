package single

import (
	"errors"
	"testing"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

func newTestModule(t *testing.T) *Module[string] {
	t.Helper()
	m := New(Lifecycle[string]{})
	m.App = service.App{DI: zdi.New(), Conf: new(service.Conf), Log: zlog.New("")}
	return m
}

func TestName(t *testing.T) {
	if got := New(Lifecycle[string]{}).Name(); got != "Single" {
		t.Fatalf("default Name() = %q, want %q", got, "Single")
	}
	if got := New(Lifecycle[string]{Name: "Worker"}).Name(); got != "Single(Worker)" {
		t.Fatalf("named Name() = %q, want %q", got, "Single(Worker)")
	}
}

func TestLifecycleCallbacks(t *testing.T) {
	calls := make([]string, 0, 3)
	m := New(Lifecycle[string]{
		Name: "Worker",
		Load: func(zdi.Invoker) (any, error) {
			calls = append(calls, "load")
			return nil, nil
		},
		Start: func(zdi.Invoker) error {
			calls = append(calls, "start")
			return nil
		},
		Done: func(zdi.Invoker) (string, error) {
			calls = append(calls, "done")
			return "ready", nil
		},
	})
	m.App = service.App{DI: zdi.New(), Conf: new(service.Conf), Log: zlog.New("")}

	if got := m.Instance(); got != "" {
		t.Fatalf("Instance() before Done = %q, want zero value", got)
	}

	if _, err := m.Load(m.DI); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := m.Start(m.DI); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := m.Done(m.DI); err != nil {
		t.Fatalf("Done: %v", err)
	}
	if got := m.Instance(); got != "ready" {
		t.Fatalf("Instance() after Done = %q, want %q", got, "ready")
	}

	want := []string{"load", "start", "done"}
	if len(calls) != len(want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("calls = %v, want %v", calls, want)
		}
	}

	if tasks := m.Tasks(); len(tasks) != 0 {
		t.Fatalf("Tasks() = %v, want empty", tasks)
	}
	if controllers := m.Controller(); len(controllers) != 0 {
		t.Fatalf("Controller() = %v, want empty", controllers)
	}
}

func TestNilCallbacksAreOptional(t *testing.T) {
	m := newTestModule(t)

	value, err := m.Load(m.DI)
	if err != nil || value != nil {
		t.Fatalf("Load with nil callback = %v, %v; want nil, nil", value, err)
	}
	if err := m.Start(m.DI); err != nil {
		t.Fatalf("Start with nil callback: %v", err)
	}
	if err := m.Done(m.DI); err != nil {
		t.Fatalf("Done with nil callback: %v", err)
	}
	if got := m.Instance(); got != "" {
		t.Fatalf("Instance() with nil Done callback = %q, want zero value", got)
	}
}

func TestDoneErrorIsReturned(t *testing.T) {
	boom := errors.New("boom")
	m := New(Lifecycle[string]{
		Done: func(zdi.Invoker) (string, error) { return "", boom },
	})
	m.App = service.App{DI: zdi.New(), Conf: new(service.Conf), Log: zlog.New("")}

	if err := m.Done(m.DI); !errors.Is(err, boom) {
		t.Fatalf("Done error = %v, want boom", err)
	}
}

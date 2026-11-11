package service

import (
	"strings"
	"testing"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zlog"
)

func TestInitTaskReportsTaskName(t *testing.T) {
	app := &App{DI: zdi.New(), Conf: &Conf{}, Log: zlog.New("")}
	tasks := []Task{
		{Name: "broken", Cron: "not a cron expression", Run: func() {}},
	}

	err := InitTask(&tasks, app)
	if err == nil {
		t.Fatal("expected an error for an invalid cron expression")
	}

	msg := strings.Join(zerror.UnwrapErrors(err), " ")
	if !strings.Contains(msg, "broken") {
		t.Fatalf("error should name the task, got: %v", err)
	}
}

func TestInitTaskSkipsIncompleteTasks(t *testing.T) {
	app := &App{DI: zdi.New(), Conf: &Conf{}, Log: zlog.New("")}

	// Tasks without a cron expression or run function are ignored, and an
	// empty task list is a no-op.
	if err := InitTask(&[]Task{}, app); err != nil {
		t.Fatalf("empty task list: %v", err)
	}

	tasks := []Task{{Name: "no-cron", Run: func() {}}, {Name: "no-run", Cron: "* * * * * *"}}
	if err := InitTask(&tasks, app); err != nil {
		t.Fatalf("incomplete tasks should be skipped: %v", err)
	}
}

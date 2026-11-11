package app_core

import (
	"errors"
	"testing"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

func TestGlobalAccessRetriesAfterInitialization(t *testing.T) {
	oldGlobal := service.Global
	service.Global = nil
	t.Cleanup(func() { service.Global = oldGlobal })

	if _, err := Conf(); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Conf before init: got %v, want ErrNotInitialized", err)
	}
	if _, err := DI(); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("DI before init: got %v, want ErrNotInitialized", err)
	}
	if _, err := Log(); !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("Log before init: got %v, want ErrNotInitialized", err)
	}

	// An early failed lookup must not be cached: after NewApp sets Global,
	// later calls have to succeed.
	service.Global = &service.App{
		DI:   zdi.New(),
		Conf: &service.Conf{},
		Log:  zlog.New(""),
	}
	if _, err := Conf(); err != nil {
		t.Fatalf("Conf after init: %v", err)
	}
	if _, err := DI(); err != nil {
		t.Fatalf("DI after init: %v", err)
	}
	if _, err := Log(); err != nil {
		t.Fatalf("Log after init: %v", err)
	}
}

func TestGlobalAccessRejectsPartiallyInitializedApp(t *testing.T) {
	oldGlobal := service.Global
	t.Cleanup(func() { service.Global = oldGlobal })

	// A Global value with nil fields must not be reported as usable: callers
	// would otherwise receive a nil pointer together with a nil error.
	service.Global = &service.App{}

	if conf, err := Conf(); !errors.Is(err, ErrNotInitialized) || conf != nil {
		t.Fatalf("Conf: got %v, %v; want nil, ErrNotInitialized", conf, err)
	}
	if di, err := DI(); !errors.Is(err, ErrNotInitialized) || di != nil {
		t.Fatalf("DI: got %v, %v; want nil, ErrNotInitialized", di, err)
	}
	if log, err := Log(); !errors.Is(err, ErrNotInitialized) || log != nil {
		t.Fatalf("Log: got %v, %v; want nil, ErrNotInitialized", log, err)
	}
}

// TestGlobalAccessAfterNewApp exercises the real initialization path instead of
// a hand-built service.App, so the nil checks in Conf/DI/Log are validated
// against what NewApp actually populates.
func TestGlobalAccessAfterNewApp(t *testing.T) {
	oldGlobal := service.Global
	service.Global = nil
	t.Cleanup(func() { service.Global = oldGlobal })

	app := service.NewApp()(zdi.New())
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.DI == nil || app.Conf == nil || app.Log == nil {
		t.Fatalf("NewApp left incomplete App: DI=%v Conf=%v Log=%v", app.DI, app.Conf, app.Log)
	}

	conf, err := Conf()
	if err != nil {
		t.Fatalf("Conf after NewApp: %v", err)
	}
	if conf != app.Conf {
		t.Fatal("Conf() returned a different instance than NewApp")
	}

	di, err := DI()
	if err != nil {
		t.Fatalf("DI after NewApp: %v", err)
	}
	if di != app.DI {
		t.Fatal("DI() returned a different instance than NewApp")
	}

	log, err := Log()
	if err != nil {
		t.Fatalf("Log after NewApp: %v", err)
	}
	if log != app.Log {
		t.Fatal("Log() returned a different instance than NewApp")
	}
}

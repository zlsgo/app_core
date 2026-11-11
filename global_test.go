package app_core

import (
	"errors"
	"testing"

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
	service.Global = &service.App{}
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

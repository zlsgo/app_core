package app_core

import (
	"errors"
	"sync"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

var (
	// ErrNotInitialized is returned when global service is not initialized
	ErrNotInitialized = errors.New("service not fully initialized")
	initOnce          sync.Once
	initError         error
)

// ensureInitialized checks if the global service is initialized
func ensureInitialized() error {
	initOnce.Do(func() {
		if service.Global == nil {
			initError = ErrNotInitialized
		}
	})
	return initError
}

func DI() (zdi.Invoker, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}
	return service.Global.DI, nil
}

// MustDI returns the DI invoker or panics if not initialized
// Deprecated: Use DI() instead and handle the error properly
func MustDI() zdi.Invoker {
	di, err := DI()
	if err != nil {
		// Keep panic for backward compatibility
		panic(err)
	}
	return di
}

func Conf() (*service.Conf, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}
	return service.Global.Conf, nil
}

// MustConf returns the configuration or panics if not initialized
// Deprecated: Use Conf() instead and handle the error properly
func MustConf() *service.Conf {
	conf, err := Conf()
	if err != nil {
		// Keep panic for backward compatibility
		panic(err)
	}
	return conf
}

func Log() (*zlog.Logger, error) {
	if err := ensureInitialized(); err != nil {
		return nil, err
	}
	return service.Global.Log, nil
}

// MustLog returns the logger or panics if not initialized
// Deprecated: Use Log() instead and handle the error properly
func MustLog() *zlog.Logger {
	log, err := Log()
	if err != nil {
		// Keep panic for backward compatibility
		panic(err)
	}
	return log
}

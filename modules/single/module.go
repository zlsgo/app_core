package single

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/zlsgo/app_core/service"
)

type Module[T any] struct {
	service.App
	lifecycle Lifecycle[T]
	instance  T
}

var (
	_ service.Module = &Module[any]{}
	_                = reflect.TypeOf(&Module[any]{})
)

func (m *Module[T]) Name() string {
	if m.lifecycle.Name != "" {
		return "Single(" + m.lifecycle.Name + ")"
	}
	return "Single"
}

func (m *Module[T]) Tasks() []service.Task {
	tasks := make([]service.Task, 0)
	return tasks
}

func (m *Module[T]) Load(zdi.Invoker) (any, error) {
	if m.lifecycle.Load != nil {
		return m.lifecycle.Load(m.DI)
	}
	return nil, nil
}

func (m *Module[T]) Start(zdi.Invoker) (err error) {
	if m.lifecycle.Start != nil {
		return m.lifecycle.Start(m.DI)
	}
	return
}

func (m *Module[T]) Done(zdi.Invoker) (err error) {
	if m.lifecycle.Done != nil {
		m.instance, err = m.lifecycle.Done(m.DI)
		return
	}
	return
}

func (m *Module[T]) Controller() []service.Controller {
	controllers := make([]service.Controller, 0)
	return controllers
}

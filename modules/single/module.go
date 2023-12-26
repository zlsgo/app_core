package single

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/zlsgo/app_core/service"
)

type Module struct {
	service.App
	lifecycle Lifecycle
	instance  interface{}
}

var (
	_ service.Module = &Module{}
	_                = reflect.TypeOf(&Module{})
)

func (m *Module) Name() string {
	if m.lifecycle.Name != "" {
		return "Single(" + m.lifecycle.Name + ")"
	}
	return "Single"
}

func (m *Module) Tasks() []service.Task {
	tasks := make([]service.Task, 0)
	return tasks
}

func (m *Module) Load(zdi.Invoker) (any, error) {
	if m.lifecycle.Load != nil {
		return m.lifecycle.Load(m.DI)
	}
	return nil, nil
}

func (m *Module) Start(zdi.Invoker) (err error) {
	if m.lifecycle.Start != nil {
		return m.lifecycle.Start(m.DI)
	}
	return
}

func (m *Module) Done(zdi.Invoker) (err error) {
	if m.lifecycle.Done != nil {
		m.instance, err = m.lifecycle.Done(m.DI)
		return
	}
	return
}

func (m *Module) Controller() []service.Controller {
	controllers := make([]service.Controller, 0)
	return controllers
}

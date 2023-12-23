package single

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/zlsgo/app_core/service"
)

type Plugin struct {
	service.App
	lifecycle Lifecycle
	instance  interface{}
}

var (
	_ service.Module = &Plugin{}
	_                = reflect.TypeOf(&Plugin{})
)

func (p *Plugin) Name() string {
	if p.lifecycle.Name != "" {
		return "Single(" + p.lifecycle.Name + ")"
	}
	return "Single"
}

func (p *Plugin) Tasks() []service.Task {
	tasks := make([]service.Task, 0)
	return tasks
}

func (p *Plugin) Load(zdi.Invoker) (any, error) {
	if p.lifecycle.Load != nil {
		return p.lifecycle.Load(p.DI)
	}
	return nil, nil
}

func (p *Plugin) Start(zdi.Invoker) (err error) {
	if p.lifecycle.Start != nil {
		return p.lifecycle.Start(p.DI)
	}
	return
}

func (p *Plugin) Done(zdi.Invoker) (err error) {
	if p.lifecycle.Done != nil {
		p.instance, err = p.lifecycle.Done(p.DI)
		return
	}
	return
}

func (p *Plugin) Controller() []service.Controller {
	controllers := make([]service.Controller, 0)
	return controllers
}

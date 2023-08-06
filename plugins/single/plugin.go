package single

import (
	"reflect"

	"github.com/zlsgo/app_core/service"
)

type Plugin struct {
	service.App
	lifecycle Lifecycle
	instance  interface{}
}

var (
	_ service.Plugin = &Plugin{}
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

func (p *Plugin) Load() (err error) {
	if p.lifecycle.Load != nil {
		return p.lifecycle.Load(p.DI)
	}
	return
}

func (p *Plugin) Start() (err error) {
	if p.lifecycle.Start != nil {
		return p.lifecycle.Start(p.DI)
	}
	return
}

func (p *Plugin) Done() (err error) {
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

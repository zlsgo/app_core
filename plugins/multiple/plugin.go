package multiple

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/zlsgo/app_core/service"
)

type Plugin struct {
	service.App
}

var (
	_ service.Plugin = &Plugin{}
	_                = reflect.TypeOf(&Plugin{})
)

func (p *Plugin) Name() string {
	names := make([]string, 0)
	plugins.ForEach(func(_ string, plugin service.Plugin) bool {
		names = append(names, plugin.Name())
		return true
	})
	if len(names) == 0 {
		return "Multiple"
	}
	return "Multiple [" + strings.Join(names, ", ") + "]"
}

func (p *Plugin) Tasks() []service.Task {
	tasks := make([]service.Task, 0)
	plugins.ForEach(func(_ string, plugin service.Plugin) bool {
		tasks = append(tasks, plugin.Tasks()...)
		return true
	})
	return tasks
}

func (p *Plugin) Load() (err error) {
	plugins.ForEach(func(_ string, plugin service.Plugin) bool {
		pdi := reflect.Indirect(reflect.ValueOf(plugin)).FieldByName("DI")
		if pdi.IsValid() {
			pdi.Set(reflect.ValueOf(p.DI))
		}
		return true
	})

	plugins.ForEach(func(_ string, plugin service.Plugin) bool {
		if e := zerror.TryCatch(plugin.Load); err != nil {
			if e != nil {
				err = zerror.With(e, plugin.Name()+" load error")
				return false
			}
		}
		return true
	})
	return
}

func (p *Plugin) Start() (err error) {
	plugins.ForEach(func(_ string, plugin service.Plugin) bool {
		if e := zerror.TryCatch(plugin.Start); err != nil {
			if e != nil {
				err = zerror.With(e, plugin.Name()+" start error")
				return false
			}
		}
		return true
	})
	return
}

func (p *Plugin) Done() (err error) {
	plugins.ForEach(func(_ string, plugin service.Plugin) bool {
		if e := zerror.TryCatch(plugin.Done); err != nil {
			if e != nil {
				err = zerror.With(e, plugin.Name()+" done error")
				return false
			}
		}
		return true
	})
	return
}

func (p *Plugin) Controller() []service.Controller {
	controllers := make([]service.Controller, 0)
	plugins.ForEach(func(_ string, plugin service.Plugin) bool {
		controllers = append(controllers, plugin.Controller()...)
		return true
	})
	return controllers
}

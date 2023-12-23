package multiple

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/zlsgo/app_core/service"
)

type Plugin struct {
	service.App
}

var (
	_ service.Module = &Plugin{}
	_                = reflect.TypeOf(&Plugin{})
)

func (p *Plugin) Name() string {
	names := make([]string, 0)
	plugins.ForEach(func(_ string, plugin service.Module) bool {
		if plugin.Name() == "" {
			return true
		}
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
	plugins.ForEach(func(_ string, plugin service.Module) bool {
		tasks = append(tasks, plugin.Tasks()...)
		return true
	})
	return tasks
}

func (p *Plugin) Load(zdi.Invoker) (any, error) {
	plugins.ForEach(func(_ string, plugin service.Module) bool {
		pdi := reflect.Indirect(reflect.ValueOf(plugin)).FieldByName("DI")
		if pdi.IsValid() {
			pdi.Set(reflect.ValueOf(p.DI))
		}
		return true
	})

	var err error
	plugins.ForEach(func(_ string, plugin service.Module) bool {
		v, e := plugin.Load(p.DI)
		if e != nil {
			err = zerror.With(e, plugin.Name()+" load error")
			return false
		}
		val := zreflect.ValueOf(v)

		if val.IsValid() {
			di := p.DI.(zdi.TypeMapper)
			if val.Kind() == reflect.Func {
				di.Provide(v)
			} else {
				di.Map(val)
			}

		}
		return true
	})
	return nil, err
}

func (p *Plugin) Start(zdi.Invoker) (err error) {
	plugins.ForEach(func(_ string, plugin service.Module) bool {
		if e := plugin.Start(p.DI); err != nil {
			if e != nil {
				err = zerror.With(e, plugin.Name()+" start error")
				return false
			}
		}
		return true
	})
	return
}

func (p *Plugin) Done(zdi.Invoker) (err error) {
	plugins.ForEach(func(_ string, plugin service.Module) bool {
		if e := plugin.Done(p.DI); err != nil {
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
	plugins.ForEach(func(_ string, plugin service.Module) bool {
		controllers = append(controllers, plugin.Controller()...)
		return true
	})
	return controllers
}

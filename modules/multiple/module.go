package multiple

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/zlsgo/app_core/service"
)

type Module struct {
	service.App
}

var (
	_ service.Module = &Module{}
	_                = reflect.TypeOf(&Module{})
)

func (m *Module) Name() string {
	names := make([]string, 0)
	modules.ForEach(func(_ string, mod service.Module) bool {
		if mod.Name() == "" {
			return true
		}
		names = append(names, mod.Name())
		return true
	})
	if len(names) == 0 {
		return "Multiple"
	}
	return "Multiple [" + strings.Join(names, ", ") + "]"
}

func (m *Module) Tasks() []service.Task {
	tasks := make([]service.Task, 0)
	modules.ForEach(func(_ string, mod service.Module) bool {
		tasks = append(tasks, mod.Tasks()...)
		return true
	})
	return tasks
}

func (m *Module) Load(zdi.Invoker) (any, error) {
	modules.ForEach(func(_ string, mod service.Module) bool {
		pdi := reflect.Indirect(reflect.ValueOf(mod)).FieldByName("DI")
		if pdi.IsValid() {
			pdi.Set(reflect.ValueOf(m.DI))
		}
		return true
	})

	var err error
	modules.ForEach(func(name string, mod service.Module) bool {
		if err = service.Utils.LoadModule(m.DI.(zdi.Injector), name, mod); err != nil {
			return false
		}
		return true
	})
	return nil, err
}

func (m *Module) Start(zdi.Invoker) (err error) {
	modules.ForEach(func(_ string, mod service.Module) bool {
		if e := mod.Start(m.DI); err != nil {
			if e != nil {
				err = zerror.With(e, mod.Name()+" start error")
				return false
			}
		}
		return true
	})
	return
}

func (m *Module) Done(zdi.Invoker) (err error) {
	modules.ForEach(func(_ string, mod service.Module) bool {
		if e := mod.Done(m.DI); e != nil {
			err = zerror.With(e, mod.Name()+" done error")
			return false
		}
		return true
	})
	return
}

func (m *Module) Controller() []service.Controller {
	controllers := make([]service.Controller, 0)
	modules.ForEach(func(_ string, mod service.Module) bool {
		controllers = append(controllers, mod.Controller()...)
		return true
	})
	return controllers
}

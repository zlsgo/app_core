package service

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
)

type Module interface {
	Name() string
	Tasks() []Task
	Controller() []Controller
	Load(zdi.Invoker) (any, error)
	Start(zdi.Invoker) error
	Done(zdi.Invoker) error
}

type ModuleService struct {
	Controllers []Controller
	Tasks       []Task
}

type ModuleLifeCycle struct {
	OnLoad  func(zdi.Invoker) (any, error)
	OnStart func(zdi.Invoker) error
	OnDone  func(zdi.Invoker) error
	OnStop  func(zdi.Invoker) error
	Service *ModuleService
}

func (p *ModuleLifeCycle) Name() string {
	return ""
}

func (p *ModuleLifeCycle) Tasks() []Task {
	if p.Service == nil {
		return nil
	}
	return p.Service.Tasks
}

func (p *ModuleLifeCycle) Controller() []Controller {
	if p.Service == nil {
		return nil
	}
	return p.Service.Controllers
}

func (p *ModuleLifeCycle) Load(di zdi.Invoker) (any, error) {
	if p.OnLoad == nil {
		return nil, nil
	}
	return p.OnLoad(di)
}

func (p *ModuleLifeCycle) Start(di zdi.Invoker) error {
	if p.OnStart == nil {
		return nil
	}
	return p.OnStart(di)
}

func (p *ModuleLifeCycle) Done(di zdi.Invoker) error {
	if p.OnDone == nil {
		return nil
	}
	return p.OnDone(di)
}

func (p *ModuleLifeCycle) Stop(di zdi.Invoker) error {
	if p.OnStop == nil {
		return nil
	}
	return p.OnStop(di)
}

// InitModule initializes the module with the given list of plugins and a dependency injector.
func InitModule(modules []Module, app *App, di zdi.Injector) (err error) {
	for _, mod := range modules {
		value := zreflect.ValueOf(mod)
		assignApp(value, app)
		_ = assignDI(value, di)
		_ = assignConf(value, app.Conf)
		name := getModuleName(mod, value)
		_ = assignLog(value, app, "[Module "+name+"] ")
		di.Map(mod)
	}

	return di.InvokeWithErrorOnly(func(app *App, tasks *[]Task, controller *[]Controller, r *Web) error {
		runs := make([]func() error, 0, len(modules))
		starts := make([]func() error, 0, len(modules))
		for i := range modules {
			mod := modules[i]
			name := getModuleName(mod, zreflect.ValueOf(mod))

			logname := zlog.Log.ColorTextWrap(zlog.ColorLightGreen, zlog.OpTextWrap(zlog.OpBold, name))
			PrintLog("Module Load", logname)

			if err := loadModule(di, name, mod); err != nil {
				return err
			}

			starts = append(starts, func() error {
				// PrintLog("Module Start", zlog.Log.ColorTextWrap(zlog.ColorLightGreen, name))
				if err := zerror.TryCatch(func() error { return mod.Start(di) }); err != nil {
					return zerror.With(err, name+" failed to Start")
				}
				return nil
			})

			runs = append(runs, func() error {
				tasks := mod.Tasks()
				if err = InitTask(&tasks); err != nil {
					return zerror.With(err, "timed task launch failed")
				}

				if err = initRouter(app, r, mod.Controller()); err != nil {
					return zerror.With(err, "router binding failed")
				}

				if err := zerror.TryCatch(func() error { return mod.Done(di) }); err != nil {
					return zerror.With(err, name+" failed to Done")
				}

				PrintLog("Module Success", logname)

				return nil
			})
		}

		for i := range starts {
			if err := starts[i](); err != nil {
				return err
			}
		}

		for i := range runs {
			if err := runs[i](); err != nil {
				return err
			}
		}
		return nil
	})
}

func getModuleName(m Module, val reflect.Value) string {
	name := m.Name()
	if name == "" {
		name = zstring.Ucfirst(strings.SplitN(val.Type().String()[1:], ".", 2)[0])
	}
	return name
}

func loadModule(di zdi.Injector, name string, mod Module) error {
	load, err := mod.Load(di)
	if err != nil {
		return zerror.With(err, name+" failed to Load")
	}

	loadVal := zreflect.ValueOf(load)
	if loadVal.IsValid() {
		if loadVal.Kind() == reflect.Func {
			di.Provide(load)
		} else {
			di.Map(load)
		}
	}

	return nil
}

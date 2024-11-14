package service

import (
	"reflect"
	"sort"
	"strings"

	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"

	"github.com/fsnotify/fsnotify"
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
func InitModule(modules []Module, app *App) (err error) {
	for _, mod := range modules {
		value := zreflect.ValueOf(mod)
		assignApp(value, app)
		_ = assignDI(value, app.DI)
		_ = assignConf(value, app.Conf)
		name := getModuleName(mod, value)
		_ = assignLog(value, app, "[Module "+name+"] ")
		_ = app.DI.(zdi.TypeMapper).Map(mod)
	}

	return app.DI.InvokeWithErrorOnly(func() error {
		var (
			tasks      *[]Task
			controller *[]Controller
			web        *Web
		)
		_ = app.DI.Resolve(&tasks, &controller, &web)

		moduleTotal := len(modules)
		runs := make([]func() error, 0, moduleTotal)
		starts := make([]func() error, 0, moduleTotal)

		type module struct {
			vof   reflect.Value
			mod   Module
			index int
		}

		modulesMap := make(map[string]module, moduleTotal)
		for i := range modules {
			mod := modules[i]
			vof := zreflect.ValueOf(mod)
			name := getModuleName(mod, vof)
			if _, ok := modulesMap[name]; ok {
				zlog.Warnf("Module %s is already registered. If you need to register multiple identical modules, please use multiple.New(...)", name)
				continue
			}
			modulesMap[name] = module{
				vof:   vof,
				index: i,
				mod:   mod,
			}
		}

		moduleKeys := zarray.Keys(modulesMap)
		sort.Strings(moduleKeys)

		app.printLog("Module", "["+strings.Join(moduleKeys, ", ")+"]")
		for v := range moduleKeys {
			name := moduleKeys[v]
			mod, vof := modulesMap[name].mod, modulesMap[name].vof
			// logname := zlog.ColorTextWrap(zlog.ColorLightGreen, zlog.OpTextWrap(zlog.OpBold, name))
			// app.printLog("Module Load", logname)

			if err := loadModule(app.DI.(zdi.Injector), name, mod); err != nil {
				return err
			}

			starts = append(starts, func() error {
				// printLog("Module Start", zlog.Log.ColorTextWrap(zlog.ColorLightGreen, name))
				if err := zerror.TryCatch(func() error { return mod.Start(app.DI) }); err != nil {
					return zerror.With(err, name+" module: failed to Start")
				}
				return nil
			})

			runs = append(runs, func() error {
				tasks := mod.Tasks()
				if err = InitTask(&tasks, app); err != nil {
					return zerror.With(err, name+" module: timed task launch failed")
				}

				if web != nil {
					if err = initRouter(app, web, mod.Controller()); err != nil {
						return zerror.With(err, name+" module: init router failed")
					}
				}

				if err := zerror.TryCatch(func() error { return mod.Done(app.DI) }); err != nil {
					return zerror.With(err, name+" module: failed to Done")
				}

				// app.printLog("Module Success", logname)

				reload := vof.MethodByName("Reload")
				if reload.IsValid() && reload.Type().Kind() == reflect.Func {
					f := reload.Interface()
					app.Conf.reloads = append(app.Conf.reloads, func() error {
						err := app.DI.InvokeWithErrorOnly(f)
						if err != nil {
							return zerror.With(err, name+" failed to Reload")
						}
						return nil
					})
				}

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

		if app.Conf.cfg != nil {
			b := zutil.NewBool(false)
			app.Conf.cfg.ConfigChange(func(e fsnotify.Event) {
				if !b.CAS(false, true) {
					return
				}
				if e.Op == fsnotify.Write {
					app.Conf.autoUnmarshal()
					for _, fn := range app.Conf.reloads {
						err = app.DI.InvokeWithErrorOnly(fn)
						if err != nil {
							zlog.Error(err)
						}
					}
				}
				b.Store(false)
			})
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

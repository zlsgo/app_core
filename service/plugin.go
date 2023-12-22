package service

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zerror"
	"github.com/sohaha/zlsgo/zreflect"
	"github.com/sohaha/zlsgo/zstring"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
)

type Plugin interface {
	Name() string
	Tasks() []Task
	Controller() []Controller
	Load(zdi.Invoker) (any, error)
	Start(zdi.Invoker) error
	Done(zdi.Invoker) error
}

type PluginService struct {
	Controllers []Controller
	Tasks       []Task
}

type Pluginer struct {
	OnLoad  func(zdi.Invoker) (any, error)
	OnStart func(zdi.Invoker) error
	OnDone  func(zdi.Invoker) error
	OnStop  func(zdi.Invoker) error
	Service *PluginService
}

func (p *Pluginer) Name() string {
	return ""
}

func (p *Pluginer) Tasks() []Task {
	if p.Service == nil {
		return nil
	}
	return p.Service.Tasks
}

func (p *Pluginer) Controller() []Controller {
	if p.Service == nil {
		return nil
	}
	return p.Service.Controllers
}

func (p *Pluginer) Load(di zdi.Invoker) (any, error) {
	if p.OnLoad == nil {
		return nil, nil
	}
	return p.OnLoad(di)
}

func (p *Pluginer) Start(di zdi.Invoker) error {
	if p.OnStart == nil {
		return nil
	}
	return p.OnStart(di)
}

func (p *Pluginer) Done(di zdi.Invoker) error {
	if p.OnDone == nil {
		return nil
	}
	return p.OnDone(di)
}

func (p *Pluginer) Stop(di zdi.Invoker) error {
	if p.OnStop == nil {
		return nil
	}
	return p.OnStop(di)
}

// InitPlugin initializes the plugin with the given list of plugins and a dependency injector.
func InitPlugin(ps []Plugin, app *App, di zdi.Injector) (err error) {
	for _, p := range ps {
		value := zreflect.ValueOf(p)
		assignApp(value, app)
		_ = assignDI(value, di)
		_ = assignConf(value, app.Conf)
		name := getPluginName(p, value)
		_ = assignLog(value, app, "[Plugin "+name+"] ")
		di.Map(p)
	}

	return di.InvokeWithErrorOnly(func(app *App, tasks *[]Task, controller *[]Controller, r *Web) error {
		runs := make([]func() error, 0, len(ps))
		starts := make([]func() error, 0, len(ps))
		for i := range ps {
			p := ps[i]
			name := getPluginName(p, zreflect.ValueOf(p))

			PrintLog("Plugin", zlog.Log.ColorTextWrap(zlog.ColorLightGreen, name))

			load, err := p.Load(di)
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

			starts = append(starts, func() error {
				if err := zerror.TryCatch(func() error { return p.Start(di) }); err != nil {
					return zerror.With(err, name+" failed to Start")
				}
				return nil
			})

			runs = append(runs, func() error {
				tasks := p.Tasks()
				if err = InitTask(&tasks); err != nil {
					return zerror.With(err, "timed task launch failed")
				}

				if err = initRouter(app, r, p.Controller()); err != nil {
					return zerror.With(err, "router binding failed")
				}

				if err := zerror.TryCatch(func() error { return p.Done(di) }); err != nil {
					return zerror.With(err, name+" failed to Done")
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
		return nil
	})
}

func getPluginName(p Plugin, val reflect.Value) string {
	name := p.Name()
	if name == "" {
		name = zstring.Ucfirst(strings.SplitN(val.Type().String()[1:], ".", 2)[0])
	}
	return name
}

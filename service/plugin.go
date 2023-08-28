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
	Load() error
	Start() error
	Done() error
}

type PluginService struct {
	Controllers []Controller
	Tasks       []Task
}

type Pluginer struct {
	OnLoad  func() error
	OnStart func() error
	OnDone  func() error
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

func (p *Pluginer) Load() error {
	if p.OnLoad == nil {
		return nil
	}
	return p.OnLoad()
}

func (p *Pluginer) Start() error {
	if p.OnStart == nil {
		return nil
	}
	return p.OnStart()
}

func (p *Pluginer) Done() error {
	if p.OnDone == nil {
		return nil
	}
	return p.OnDone()
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

	return di.InvokeWithErrorOnly(func(tasks *[]Task, controller *[]Controller, r *Web) error {
		start := make([]func() error, 0, len(ps))
		for i := range ps {
			p := ps[i]
			name := getPluginName(p, zreflect.ValueOf(p))

			err := p.Load()
			if err != nil {
				return zerror.With(err, name+" failed to Load")
			}

			PrintLog("Plugin", zlog.Log.ColorTextWrap(zlog.ColorLightGreen, name))

			err = p.Start()
			if err != nil {
				return zerror.With(err, name+" failed to Start")
			}

			start = append(start, func() error {
				*tasks = append(*tasks, p.Tasks()...)
				*controller = append(*controller, p.Controller()...)
				if err := zerror.TryCatch(func() error { return p.Done() }); err != nil {
					return zerror.With(err, name+" failed to Done")
				}
				return nil
			})
		}

		for i := range start {
			if err := start[i](); err != nil {
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

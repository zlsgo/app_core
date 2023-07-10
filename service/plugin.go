package service

import (
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zerror"
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

// InitPlugin initializes the plugin with the given list of plugins and a dependency injector.
func InitPlugin(ps []Plugin, di zdi.Injector) (err error) {
	for _, p := range ps {
		pdi := reflect.Indirect(reflect.ValueOf(p)).FieldByName("DI")
		if pdi.IsValid() {
			// switch pdi.Type().String() {
			// case "zdi.Invoker", "zdi.Injector":
			pdi.Set(reflect.ValueOf(di))
			// }
		}
		di.Map(p)
	}

	return di.InvokeWithErrorOnly(func(app *App, tasks *[]Task, controller *[]Controller, r *Web) error {
		start := make([]func() error, 0, len(ps))
		for i := range ps {
			p := ps[i]
			val := reflect.ValueOf(p)
			name := p.Name()
			if name == "" {
				name = zstring.Ucfirst(strings.SplitN(val.Type().String()[1:], ".", 2)[0])
			}

			ival := reflect.Indirect(val)
			conf := ival.FieldByName("Conf")
			if conf.IsValid() && conf.Type().String() == "*service.Conf" {
				conf.Set(reflect.ValueOf(app.Conf))
			}

			log := ival.FieldByName("Log")
			if log.IsValid() && log.Type().String() == "*zlog.Logger" {
				log.Set(reflect.ValueOf(app.Log))
			}

			err := zerror.TryCatch(func() error {
				return p.Load()
			})
			if err != nil {
				return zerror.With(err, name+" failed to Load")
			}

			PrintLog("Plugin", zlog.Log.ColorTextWrap(zlog.ColorLightGreen, name))
			err = zerror.TryCatch(func() error {
				return p.Start()
			})
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

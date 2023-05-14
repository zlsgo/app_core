package service

import (
	"reflect"
	"strings"

	"github.com/zlsgo/app_core/utils"

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

func InitPlugin(ps []Plugin, di zdi.Injector) (err error) {
	for _, p := range ps {
		pdi := reflect.Indirect(reflect.ValueOf(p)).FieldByName("DI")
		if pdi.IsValid() {
			switch pdi.Type().String() {
			case "zdi.Invoker", "zdi.Injector":
				pdi.Set(reflect.ValueOf(di))
			}
		}

		di.Map(p)
	}

	return utils.InvokeErr(di.Invoke(func(app *App, tasks *[]Task, controller *[]Controller, r *Web) error {
		start := make([]func() error, 0, len(ps))
		for i := range ps {
			p := ps[i]
			val := reflect.ValueOf(p)
			name := p.Name()
			if name == "" {
				name = zstring.Ucfirst(strings.SplitN(val.Type().String()[1:], ".", 2)[0])
			}

			*tasks = append(*tasks, p.Tasks()...)
			*controller = append(*controller, p.Controller()...)

			conf := reflect.Indirect(val).FieldByName("Conf")
			if conf.IsValid() && conf.Type().String() == "*service.Conf" {
				conf.Set(reflect.ValueOf(app.Conf))
			}

			log := reflect.Indirect(val).FieldByName("Log")
			if log.IsValid() && log.Type().String() == "*zlog.Logger" {
				log.Set(reflect.ValueOf(app.Log))
			}

			err := zerror.TryCatch(func() error {
				return p.Load()
			})
			if err != nil {
				return zerror.With(err, name)
			}

			PrintLog("Plugin", zlog.Log.ColorTextWrap(zlog.ColorLightGreen, name))
			err = zerror.TryCatch(func() error {
				return p.Start()
			})

			start = append(start, func() error {

				if err := zerror.TryCatch(func() error { return p.Done() }); err != nil {
					return zerror.With(err, name)
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
	}))
}

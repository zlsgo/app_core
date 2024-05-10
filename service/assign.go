package service

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
)

func dynamicAssign(value reflect.Value, app *App) (err error) {
	assignApp(value, app)
	err = assignDI(value, app.DI)
	if err != nil {
		return err
	}
	err = assignConf(value, app.Conf)
	if err != nil {
		return err
	}
	err = assignLog(value, app, "")
	return
}

func assignDI(value reflect.Value, di zdi.Invoker) error {
	if value.Kind() != reflect.Ptr {
		return nil
	}
	e := value.Elem()
	for _, d := range []string{"DI", "di", "Di"} {
		cDI := e.FieldByName(d)
		if cDI.IsValid() {
			switch cDI.Type().String() {
			case "zdi.Invoker", "zdi.Injector":
				if err := zreflect.SetUnexportedField(value, d, di); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func assignConf(value reflect.Value, conf *Conf) error {
	if value.Kind() != reflect.Ptr {
		return nil
	}
	e := value.Elem()
	for _, d := range []string{"conf", "Conf"} {
		v := e.FieldByName(d)
		if v.IsValid() && v.Type().String() == "*service.Conf" {
			if err := zreflect.SetUnexportedField(value, d, conf); err != nil {
				return err
			}
		}
	}
	return nil
}

func assignLog(value reflect.Value, app *App, name string) error {
	if value.Kind() != reflect.Ptr {
		return nil
	}
	e := value.Elem()
	for _, d := range []string{"log", "Log"} {
		v := e.FieldByName(d)
		if v.IsValid() && v.Type().String() == "*zlog.Logger" {
			pLog := zlog.New(name)
			pLog.ResetFlags(app.Log.GetFlags())
			pLog.SetLogLevel(app.Log.GetLogLevel())
			// pLog.ResetWriter(app.Log.GetWriter())
			if err := zreflect.SetUnexportedField(value, d, pLog); err != nil {
				return err
			}
		}
	}
	return nil
}

func assignApp(value reflect.Value, app *App) {
	valueOf := reflect.Indirect(value)
	api := -1
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).Type().Name() == "App" {
			api = i
			break
		}
	}

	if api != -1 {
		valueOf.Field(api).Set(reflect.ValueOf(*app))
		// return fmt.Errorf("%s not a legitimate controller", controller)
	}
}

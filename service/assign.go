package service

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zreflect"
)

func dynamicAssign(value reflect.Value, app *App) (err error) {
	if !value.IsValid() || value.Kind() != reflect.Ptr || value.IsNil() {
		return nil
	}
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

// AssignApp injects the application context (App, DI, Conf and Log) into value
// when it has matching fields. Composite modules use it to initialize their
// child modules before running their lifecycle callbacks.
func AssignApp(value interface{}, app *App) error {
	return dynamicAssign(reflect.ValueOf(value), app)
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
			pLog.Writer().Reset(app.Log)
			if err := zreflect.SetUnexportedField(value, d, pLog); err != nil {
				return err
			}
		}
	}
	return nil
}

func assignApp(value reflect.Value, app *App) {
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return
	}

	valueOf := value.Elem()
	appType := reflect.TypeOf(App{})
	for i := 0; i < valueOf.NumField(); i++ {
		if valueOf.Field(i).Type() != appType {
			continue
		}
		_ = zreflect.SetUnexportedField(value, valueOf.Type().Field(i).Name, *app)
		return
	}
}

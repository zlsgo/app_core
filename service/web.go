package service

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/zreflect"
	"github.com/zlsgo/app_core/common"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zpprof"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
)

type (
	// Web represents a web structure.
	Web struct {
		*znet.Engine
		hijacked []func(c *znet.Context) bool
	}

	// Controller is an interface for controller functions.
	Controller interface {
		Init(r *znet.Engine) error
	}

	// RouterBeforeProcess is a type for controller pre-processing.
	RouterBeforeProcess func(r *Web, app *App)

	// Template represents a template structure.
	Template struct {
		Global ztype.Map // Global is a map for global variables in the template.
		DIR    string    // DIR is the directory path for the template.
	}
)

// AddHijack adds a hijack function to the Web struct
func (w *Web) AddHijack(fn func(c *znet.Context) bool) {
	if fn == nil {
		return
	}
	w.hijacked = append(w.hijacked, fn)
}

// GetHijack returns the hijacked functions of the Web struct
func (w *Web) GetHijack() []func(c *znet.Context) bool {
	return w.hijacked
}

// NewWeb returns a function that creates a new Web instance along with a znet.Engine instance
func NewWeb() func(app *App, middlewares []znet.Handler, plugin []Plugin) (*Web, *znet.Engine) {
	return func(app *App, middlewares []znet.Handler, ps []Plugin) (*Web, *znet.Engine) {
		r := znet.New()
		r.Log = app.Log
		r.AllowQuerySemicolons = true
		zlog.Log = r.Log

		r.BindStructSuffix = ""
		r.BindStructDelimiter = "-"
		r.SetAddr(app.Conf.Base.Port)

		isDebug := app.Conf.Base.Debug
		if isDebug {
			r.SetMode(znet.DebugMode)
		} else {
			r.SetMode(znet.ProdMode)
		}

		if app.Conf.Base.Pprof {
			zpprof.Register(r, app.Conf.Base.PprofToken)
		}

		var errHandler znet.ErrHandlerFunc
		if err := app.DI.Resolve(&errHandler); err == nil {
			r.Use(znet.RewriteErrorHandler(errHandler))
		}

		for _, middleware := range middlewares {
			r.Use(middleware)
		}

		return &Web{
			Engine: r,
		}, r
	}
}

// RunWeb runs the web application
func RunWeb(r *Web, app *App, controllers *[]Controller, p []Plugin) {
	_, err := app.DI.Invoke(func(after RouterBeforeProcess) {
		after(r, app)
	})
	if err != nil && !strings.Contains(err.Error(), "value not found for type service.RouterBeforeProcess") {
		common.Fatal(err)
	}

	common.Fatal(initRouter(app, r, *controllers))
	r.StartUp()
}

// initRouter initializes the router for the application
func initRouter(app *App, _ *Web, controllers []Controller) (err error) {
	err = app.DI.InvokeWithErrorOnly(func(r *Web) error {
		for i := range controllers {
			c := controllers[i]
			valueOf := zreflect.ValueOf(c)
			typeOf := zreflect.TypeOf(c).Elem()
			value := reflect.Indirect(valueOf)
			controller := strings.TrimPrefix(typeOf.String(), "controller.")
			controller = strings.Replace(controller, ".", "/", -1)

			err = dynamicAssign(valueOf, app)
			if err != nil {
				return err
			}

			name := getWebRouterName(value, controller)
			if name == "" {
				err = r.BindStruct(name, c)
			} else {
				err = r.Group("/").BindStruct(name, c)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		err = fmt.Errorf("初始化路由失败: %w", err)

	}
	return
}

func getWebRouterName(value reflect.Value, controller string) string {
	name := ""
	cName := value.FieldByName("Path")
	if cName.IsValid() && cName.String() != "" {
		name = zstring.CamelCaseToSnakeCase(cName.String(), "/")
	} else {
		name = zstring.CamelCaseToSnakeCase(controller, "/")
	}

	lname := strings.Split(name, "/")
	if lname[len(lname)-1] == "index" {
		name = strings.Join(lname[:len(lname)-1], "/")
		name = strings.TrimSuffix(name, "/")
	}
	return name
}

package service

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zlsgo/app_core/utils"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zpprof"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	Web struct {
		*znet.Engine
		hijacked []func(c *znet.Context) bool
	}
	// Controller 控制器函数
	Controller interface {
		Init(r *znet.Engine) error
	}
	// RouterBeforeProcess 控制器前置处理
	RouterBeforeProcess func(r *Web, app *App)
	Template            struct {
		Global ztype.Map
		DIR    string
	}
)

func (w *Web) AddHijack(fn func(c *znet.Context) bool) {
	if fn == nil {
		return
	}
	w.hijacked = append(w.hijacked, fn)
}

func (w *Web) GetHijack() []func(c *znet.Context) bool {
	return w.hijacked
}

// NewWeb 初始化 WEB
func NewWeb() func(app *App, middlewares []znet.Handler) (*Web, *znet.Engine) {
	return func(app *App, middlewares []znet.Handler) (*Web, *znet.Engine) {
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

func RunWeb(r *Web, app *App, controllers *[]Controller) {
	_, err := app.DI.Invoke(func(after RouterBeforeProcess) {
		after(r, app)
	})
	if err != nil && !strings.Contains(err.Error(), "value not found for type service.RouterBeforeProcess") {
		utils.Fatal(err)
	}

	utils.Fatal(initRouter(app, r, *controllers))
	r.StartUp()
}

func initRouter(app *App, _ *Web, controllers []Controller) (err error) {
	_, _ = app.DI.Invoke(func(r *Web) {
		for i := range controllers {
			c := controllers[i]
			err = zutil.TryCatch(func() (err error) {
				typeOf := reflect.TypeOf(c).Elem()
				controller := strings.TrimPrefix(typeOf.String(), "controller.")
				controller = strings.Replace(controller, ".", "/", -1)
				api := -1
				for i := 0; i < typeOf.NumField(); i++ {
					if typeOf.Field(i).Type.String() == "service.App" {
						api = i
						break
					}
				}
				if api == -1 {
					return fmt.Errorf("%s not a legitimate controller", controller)
				}

				reflect.ValueOf(c).Elem().Field(api).Set(reflect.ValueOf(*app))

				cDI := reflect.Indirect(reflect.ValueOf(c)).FieldByName("DI")
				if cDI.IsValid() {
					switch cDI.Type().String() {
					case "zdi.Invoker", "zdi.Injector":
						cDI.Set(reflect.ValueOf(app.DI))
					}
				}

				name := ""
				cName := reflect.Indirect(reflect.ValueOf(c)).FieldByName("Path")

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
				if name == "" {
					err = r.BindStruct(name, c)
				} else {
					err = r.Group("/").BindStruct(name, c)
				}
				return err
			})

			if err != nil {
				err = fmt.Errorf("初始化路由失败: %w", err)
				return
			}
		}
	})
	return
}

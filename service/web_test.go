package service

import (
	"reflect"
	"testing"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/znet"
)

type routeTarget struct {
	Path string
}

func TestGetWebRouterName(t *testing.T) {
	cases := []struct {
		path       string
		controller string
		want       string
	}{
		{"", "Index", ""},
		{"", "User", "user"},
		{"", "api/Index", "api"},
		{"", "api/User", "api/user"},
		{"api", "Index", "api"},
		{"api/v1", "Index", "api/v1"},
		{"AdminPanel", "Index", "admin_panel"},
	}

	for _, tc := range cases {
		value := reflect.ValueOf(&routeTarget{Path: tc.path}).Elem()
		if got := getWebRouterName(value, tc.controller); got != tc.want {
			t.Fatalf("path=%q controller=%q: got %q, want %q", tc.path, tc.controller, got, tc.want)
		}
	}
}

type probeController struct {
	Path string
}

func (c *probeController) Init(r *znet.Engine) error { return nil }

type valueController struct{}

func (valueController) Init(r *znet.Engine) error { return nil }

func TestBindControllerRejectsInvalidControllers(t *testing.T) {
	app := &App{DI: zdi.New(), Conf: &Conf{}, Log: zlog.New("")}
	r := &Web{Engine: znet.New()}

	cases := map[string]Controller{
		"nil":       nil,
		"typed nil": (*probeController)(nil),
		"value":     valueController{},
	}
	for name, c := range cases {
		if err := bindController(app, r, c); err == nil {
			t.Fatalf("%s controller: expected an error", name)
		}
	}
}

func (c *probeController) GET(ctx *znet.Context) string { return "ok" }

func TestBindControllerRegistersRoutes(t *testing.T) {
	app := &App{DI: zdi.New(), Conf: &Conf{}, Log: zlog.New("")}
	r := &Web{Engine: znet.New()}

	if err := bindController(app, r, &probeController{Path: "api"}); err != nil {
		t.Fatalf("bindController: %v", err)
	}
}

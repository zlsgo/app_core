package service

import (
	"reflect"
	"testing"
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

package multiple

import (
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

var modules = zarray.NewHashMap[string, service.Module]()

func Get(name string) (service.Module, bool) {
	return modules.Get(name)
}

func New(mods map[string]service.Module) *Module {
	if modules.Len() > 0 {
		zlog.Warn("modules already exists, please check if there are duplicate names")
	}
	for name, m := range mods {
		modules.Set(name, m)
	}
	return &Module{}
}

func (m *Module) Add(name string, plugin service.Module) {
	modules.Set(name, plugin)
}

func (m *Module) Get(name string) (service.Module, bool) {
	return Get(name)
}

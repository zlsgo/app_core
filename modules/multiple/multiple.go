package multiple

import (
	"github.com/sohaha/zlsgo/zarray"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

var plugins = zarray.NewHashMap[string, service.Module]()

func Get(name string) (service.Module, bool) {
	return plugins.Get(name)
}

func New(ps map[string]service.Module) *Plugin {
	if plugins.Len() > 0 {
		zlog.Warn("plugins already exists, please check if there are duplicate names")
	}
	for name, plugin := range ps {
		plugins.Set(name, plugin)
	}
	return &Plugin{}
}

func (p *Plugin) Add(name string, plugin service.Module) {
	plugins.Set(name, plugin)
}

func (p *Plugin) Get(name string) (service.Module, bool) {
	return Get(name)
}

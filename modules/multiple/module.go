package multiple

import (
	"reflect"
	"sort"
	"strings"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zerror"
	"github.com/zlsgo/app_core/service"
)

type Module struct {
	service.App
}

var (
	_ service.Module = &Module{}
	_                = reflect.TypeOf(&Module{})
)

type namedModule struct {
	key string
	mod service.Module
}

// sortedModules returns the registered modules ordered by their registration
// key. The underlying map has no stable iteration order, so every lifecycle
// stage walks this slice to keep behaviour reproducible.
func sortedModules() []namedModule {
	keys := modules.Keys()
	sort.Strings(keys)

	list := make([]namedModule, 0, len(keys))
	for _, key := range keys {
		mod, ok := modules.Get(key)
		if !ok || service.IsNilModule(mod) {
			continue
		}
		list = append(list, namedModule{key: key, mod: mod})
	}
	return list
}

func (m *Module) Name() string {
	names := make([]string, 0)
	for _, item := range sortedModules() {
		if item.mod.Name() == "" {
			continue
		}
		names = append(names, item.mod.Name())
	}
	if len(names) == 0 {
		return "Multiple"
	}
	return "Multiple [" + strings.Join(names, ", ") + "]"
}

func (m *Module) Tasks() []service.Task {
	tasks := make([]service.Task, 0)
	for _, item := range sortedModules() {
		tasks = append(tasks, item.mod.Tasks()...)
	}
	return tasks
}

func (m *Module) Load(zdi.Invoker) (any, error) {
	for _, item := range sortedModules() {
		if err := service.AssignApp(item.mod, &m.App); err != nil {
			return nil, zerror.With(err, item.key+" assign error")
		}
		if err := service.Utils.LoadModule(m.DI.(zdi.Injector), item.key, item.mod); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (m *Module) Start(zdi.Invoker) (err error) {
	for _, item := range sortedModules() {
		if e := item.mod.Start(m.DI); e != nil {
			return zerror.With(e, item.mod.Name()+" start error")
		}
	}
	return nil
}

func (m *Module) Done(zdi.Invoker) (err error) {
	for _, item := range sortedModules() {
		if e := item.mod.Done(m.DI); e != nil {
			return zerror.With(e, item.mod.Name()+" done error")
		}
	}
	return nil
}

func (m *Module) Controller() []service.Controller {
	controllers := make([]service.Controller, 0)
	for _, item := range sortedModules() {
		controllers = append(controllers, item.mod.Controller()...)
	}
	return controllers
}

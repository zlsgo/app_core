package single

import (
	"github.com/sohaha/zlsgo/zdi"
)

type Lifecycle struct {
	Name  string
	Load  func(zdi.Invoker) (any, error)
	Start func(zdi.Invoker) error
	Done  func(zdi.Invoker) (interface{}, error)
}

func New(s Lifecycle) *Module {
	return &Module{
		lifecycle: s,
	}
}

func (m *Module) Instance() interface{} {
	return m.instance
}

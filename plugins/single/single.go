package single

import (
	"github.com/sohaha/zlsgo/zdi"
)

type Lifecycle struct {
	Name  string
	Load  func(zdi.Invoker) error
	Start func(zdi.Invoker) error
	Done  func(zdi.Invoker) (interface{}, error)
}

func New(s Lifecycle) *Plugin {
	return &Plugin{
		lifecycle: s,
	}
}

func (p *Plugin) Instance() interface{} {
	return p.instance
}

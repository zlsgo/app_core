package single

import (
	"github.com/sohaha/zlsgo/zdi"
)

type Lifecycle[T any] struct {
	Name  string
	Load  func(zdi.Invoker) (any, error)
	Start func(zdi.Invoker) error
	Done  func(zdi.Invoker) (T, error)
}

func New[T any](s Lifecycle[T]) *Module[T] {
	return &Module[T]{
		lifecycle: s,
	}
}

func (m *Module[T]) Instance() T {
	return m.instance
}

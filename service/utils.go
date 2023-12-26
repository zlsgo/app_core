package service

import (
	"github.com/sohaha/zlsgo/zdi"
)

type (
	utils struct{}
)

var Utils = utils{}

func (utils) LoadModule(di zdi.Injector, name string, mod Module) error {
	return loadModule(di, name, mod)
}

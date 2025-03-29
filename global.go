package app_core

import (
	"errors"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

func DI() (zdi.Invoker, error) {
	if service.Global == nil {
		return nil, errors.New("not fully initialized globally")
	}

	return service.Global.DI, nil
}

func MustDI() zdi.Invoker {
	di, err := DI()
	if err != nil {
		panic(err)
	}
	return di
}

func Conf() (*service.Conf, error) {
	if service.Global == nil {
		return nil, errors.New("not fully initialized globally")
	}

	return service.Global.Conf, nil
}

func MustConf() *service.Conf {
	conf, err := Conf()
	if err != nil {
		panic(err)
	}
	return conf
}

func Log() (*zlog.Logger, error) {
	if service.Global == nil {
		return nil, errors.New("not fully initialized globally")
	}

	return service.Global.Log, nil
}

func MustLog() *zlog.Logger {
	log, err := Log()
	if err != nil {
		panic(err)
	}
	return log
}

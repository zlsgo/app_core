package multiple

import (
	"testing"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/zlsgo/app_core/service"
)

type probeModule struct {
	service.App
	gotDI   bool
	gotConf bool
	gotLog  bool
}

func (p *probeModule) Name() string                     { return "ProbeModule" }
func (p *probeModule) Tasks() []service.Task            { return nil }
func (p *probeModule) Controller() []service.Controller { return nil }
func (p *probeModule) Start(zdi.Invoker) error          { return nil }
func (p *probeModule) Done(zdi.Invoker) error           { return nil }
func (p *probeModule) Load(zdi.Invoker) (any, error) {
	p.gotDI = p.App.DI != nil
	p.gotConf = p.App.Conf != nil
	p.gotLog = p.App.Log != nil
	return nil, nil
}

func TestLoadInjectsChildModules(t *testing.T) {
	di := zdi.New()
	child := &probeModule{}
	bundle := New(map[string]service.Module{"probe": child})
	bundle.App = service.App{DI: di, Conf: new(service.Conf), Log: zlog.New("")}

	if _, err := bundle.Load(di); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !child.gotDI || !child.gotConf || !child.gotLog {
		t.Fatalf("child module not fully injected: DI=%v Conf=%v Log=%v",
			child.gotDI, child.gotConf, child.gotLog)
	}
}

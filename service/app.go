package service

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

// App represents an application.
type App struct {
	DI   zdi.Invoker  // Dependency injection invoker.
	Conf *Conf        // Application configuration.
	Log  *zlog.Logger // Logger instance.
}

var (
	_      = reflect.TypeOf(&App{})
	Global *App
)

// NewApp creates a new App with the provided options.
func NewApp(opt ...func(o *BaseConf)) func(conf *Conf, di zdi.Injector) *App {
	baseConf = zutil.Optional(baseConf, opt...)
	RegisterDefaultConf(baseConf)

	log := zlog.New(LogPrefix)
	log.ResetFlags(zlog.BitLevel | zlog.BitTime)
	zlog.SetDefault(log)

	return func(conf *Conf, di zdi.Injector) *App {
		Global = &App{
			DI:   di,
			Conf: conf,
			Log:  setLog(log, conf),
		}
		return Global
	}
}

// setLog initializes the logger with the given configuration.
func setLog(log *zlog.Logger, c *Conf) *zlog.Logger {
	if c.Base.LogPosition {
		log.ResetFlags(log.GetFlags() | zlog.BitLongFile)
	}

	if c.Base.LogDir != "" {
		log.SetSaveFile(zfile.RealPath(c.Base.LogDir, true)+"app.log", true)
	}

	if c.Base.Debug {
		log.SetLogLevel(zlog.LogDump)
	} else {
		log.SetLogLevel(zlog.LogSuccess)
	}

	return log
}

// PrintLog prints a log message with the given tip and additional values
func PrintLog(tip string, v ...interface{}) {
	d := []interface{}{
		zlog.ColorTextWrap(zlog.ColorLightMagenta, zstring.Pad(tip, 6, " ", zstring.PadLeft)),
	}
	d = append(d, v...)
	zlog.Debug(d...)
}

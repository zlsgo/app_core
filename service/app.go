package service

import (
	"reflect"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
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
	b := BaseConf{
		Debug: debug,
		Zone:  8,
		Port:  "3788",
	}
	for _, o := range opt {
		o(&b)
	}
	DefaultConf = append(DefaultConf, b)

	return func(conf *Conf, di zdi.Injector) *App {
		Global = &App{
			DI:   di,
			Conf: conf,
			Log:  initLog(conf),
		}
		return Global
	}
}

// initLog initializes the logger with the given configuration.
func initLog(c *Conf) *zlog.Logger {
	log := zlog.Log
	log.SetPrefix(LogPrefix)

	logFlags := zlog.BitLevel | zlog.BitTime
	if c.Base.LogPosition {
		logFlags = logFlags | zlog.BitLongFile
	}
	log.ResetFlags(logFlags)

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

package service

import (
	"reflect"
	"strings"

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
func NewApp(opt ...func(o BaseConf) BaseConf) func(conf *Conf, di zdi.Injector) *App {
	for i := range opt {
		baseConf = opt[i](baseConf)
	}

	RegisterDefaultConf(baseConf)

	log := zlog.New(LogPrefix)
	log.ResetFlags(zlog.BitLevel | zlog.BitTime)

	if baseConf.DisableDebug {
		log.Discard()
	}

	if !baseConf.Debug {
		log.SetLogLevel(zlog.LogSuccess)
	}

	zlog.SetDefault(log)

	return func(conf *Conf, di zdi.Injector) *App {
		Global = &App{
			DI:   di,
			Conf: conf,
			Log:  setLog(log, conf),
		}
		_ = di.Maps(di, conf, Global)
		return Global
	}
}

// setLog initializes the logger with the given configuration.
func setLog(log *zlog.Logger, c *Conf) *zlog.Logger {
	logFlags := log.GetFlags()
	if c.Base.LogPosition {
		logFlags |= zlog.BitLongFile
	}

	if baseConf.LogShowDate {
		logFlags |= zlog.BitDate
	}
	log.ResetFlags(logFlags)

	var logfile string
	if c.Base.LogDir != "" {
		if c.Base.LogFile == "" {
			c.Base.LogFile = "app.log"
		}
		logfile = c.Base.LogDir + "/" + c.Base.LogFile
	} else if c.Base.LogFile != "" {
		logfile = c.Base.LogFile
	}

	if logfile != "" {
		log.SetSaveFile(zfile.RealPath(logfile), true)
	}

	if c.Base.Debug {
		log.SetLogLevel(zlog.LogDump)
	} else {
		log.SetLogLevel(zlog.LogSuccess)
	}

	if c.Base.LogLevel != "" {
		switch strings.ToLower(c.Base.LogLevel) {
		case "debug":
			log.SetLogLevel(zlog.LogDump)
		case "info":
			log.SetLogLevel(zlog.LogSuccess)
		case "warn":
			log.SetLogLevel(zlog.LogWarn)
		}
	}

	if c.Base.LogMaxAge != 0 {
		zlog.LogMaxDurationDate = c.Base.LogMaxAge
	}

	zlog.SetDefault(log)

	return log
}

// printLog prints a log message with the given tip and additional values
func (app *App) printLog(tip string, v ...interface{}) {
	d := []interface{}{
		app.Log.ColorTextWrap(zlog.ColorLightMagenta, zstring.Pad(tip, 6, " ", zstring.PadLeft)),
	}
	d = append(d, v...)
	app.Log.Debug(d...)
}

package service

import (
	"reflect"

	"github.com/zlsgo/app_core/common"

	"github.com/sohaha/zlsgo/ztype"
	gconf "github.com/zlsgo/conf"
)

type BaseConf struct {
	// LogDir specifies the directory for log files.
	LogDir string `z:"log_dir,omitempty"`

	// Port specifies the port number for the server.
	Port string `z:"port,omitempty"`

	// PprofToken is a token for accessing pprof endpoints.
	PprofToken string `z:"pprof_token,omitempty"`

	// Zone specifies the zone for the configuration.
	Zone int8 `z:"zone,omitempty"`

	// Debug specifies if debug mode is enabled.
	Debug bool `z:"debug,omitempty"`

	// LogPosition specifies if log position should be included in logs.
	LogPosition bool `z:"log_position,omitempty"`

	// Pprof specifies if pprof endpoints are enabled.
	Pprof bool `z:"pprof,omitempty"`

	// DisableDebug specifies if debug mode is disabled.
	DisableDebug bool `z:"-"`
}

// ConfKey returns the configuration key for the BaseConf struct
func (BaseConf) ConfKey() string {
	return "base"
}

var (
	// fileName is the name of the configuration file.
	fileName = "conf"

	// LogPrefix is the prefix for log messages.
	LogPrefix = ""

	// AppName is the name of the application.
	AppName = "ZlsAPP"

	// debug determines whether debug mode is enabled.
	debug = true
)

var (
	DefaultConf []interface{}
)

// Conf represents the configuration struct.
type Conf struct {
	cfg *gconf.Confhub // cfg is used to manage the configuration settings.

	Base BaseConf // Base represents the base configuration settings.
}

// Get retrieves the value associated with the given key from the Conf object.
//
// Parameters:
// - key: the key used to identify the value to retrieve.
//
// Returns:
// - ztype.Type: the value associated with the given key.
func (c *Conf) Get(key string) ztype.Type {
	return c.cfg.Get(key)
}

// Unmarshal unmarshals the value associated with the given key in the Conf struct.
//
// It takes a string key and a pointer to an interface{} as its parameters.
// The function returns an error.
func (c *Conf) Unmarshal(key string, rawVal interface{}) error {
	return c.cfg.UnmarshalKey(key, &rawVal)
}

// NewConf creates a new Conf object with the given options.
//
// opt: The optional configuration options.
//
//	These options are functions that modify the Conf object.
//	They can be used to customize the behavior of the Conf object.
//	The functions should accept a pointer to a gconf.Options object.
//	Example:
//	func(o *gconf.Options) {
//	    o.EnvPrefix = AppName
//	    o.AutoCreate = true
//	    o.PrimaryAliss = "dev"
//	}
func NewConf(opt ...func(o *gconf.Options)) func() *Conf {
	cfg := gconf.New(fileName, func(o *gconf.Options) {
		o.EnvPrefix = AppName
		o.AutoCreate = true
		o.PrimaryAliss = "dev"
		for _, f := range opt {
			f(o)
		}
	})

	return func() *Conf {
		c := &Conf{cfg: cfg}
		delay := setConf(c, DefaultConf)
		common.Fatal(cfg.Read())
		delay()
		common.Fatal(cfg.Unmarshal(&c))

		return c
	}
}

func (c *Conf) Write() error {
	return c.cfg.Write()
}

// Set updates the value of a configuration key.
//
// key: the key of the configuration property to be updated.
// value: the new value to be set for the configuration property.
func (c *Conf) Set(key string, value interface{}) {
	c.cfg.Set(key, value)
}

func getConfName(t reflect.Value) string {
	var key string
	getConfKey := t.MethodByName("ConfKey")
	if getConfKey.IsValid() {
		g, ok := getConfKey.Interface().(func() string)
		if ok {
			key = g()
		}
	}

	if key == "" {
		if t.Kind() == reflect.Slice {
			key = t.Type().Elem().Name()
		} else {
			key = t.Type().Name()
		}
	}

	return key
}

func setConf(conf *Conf, value []interface{}) func() {
	confs, disableDebug, autoUnmarshal := ztype.Map{}, false, []func(){}
	setConf := func(disableWrite bool) func(key string, value interface{}) {
		if !disableWrite {
			return conf.cfg.SetDefault
		}
		return func(key string, value interface{}) {
			confs[key] = value
		}
	}

	for i := range value {
		c := value[i]
		v := reflect.ValueOf(c)
		d := v.MethodByName("DisableWrite")
		disableWrite := false
		if d.IsValid() {
			if f, ok := d.Interface().(func() bool); ok {
				disableWrite = f()
			}
		}
		isPtr := v.Kind() == reflect.Ptr
		name := getConfName(v)
		v = reflect.Indirect(v)
		set := setConf(disableWrite)
		t := v.Type()
		switch t.Kind() {
		case reflect.Struct:
			m := ztype.ToMap(c)
			if name == "base" {
				disableDebug = m.Get("DisableDebug").Bool()
			}
			set(name, m)
		case reflect.Slice:
			switch t.Elem().Kind() {
			case reflect.Struct:
				m := make([]map[string]interface{}, v.Len())
				for i := 0; i < v.Len(); i++ {
					m[i] = ztype.ToMap(v.Index(i).Interface())
				}
				set(name, m)
			default:
				set(name, v)
			}
		default:
			set(name, c)
		}

		if isPtr {
			autoUnmarshal = append(autoUnmarshal, func() {
				_ = conf.cfg.UnmarshalKey(name, &c)
			})
		}
	}

	if disableDebug {
		conf.Base.Debug = false
	}

	return func() {
		for k, v := range confs {
			conf.cfg.SetDefault(k, v)
		}
		for _, f := range autoUnmarshal {
			f()
		}
	}
}

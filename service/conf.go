package service

import (
	"reflect"

	"github.com/zlsgo/app_core/common"

	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/spf13/viper"
	gconf "github.com/zlsgo/conf"
)

type BaseConf struct {
	// LogDir specifies the directory for log files.
	LogDir string `mapstructure:"log_dir"`

	// Port specifies the port number for the server.
	Port string `mapstructure:"port"`

	// PprofToken is a token for accessing pprof endpoints.
	PprofToken string `mapstructure:"pprof_token"`

	// Zone specifies the zone for the configuration.
	Zone int8 `mapstructure:"zone"`

	// Debug specifies if debug mode is enabled.
	Debug bool `mapstructure:"debug"`

	// LogPosition specifies if log position should be included in logs.
	LogPosition bool `mapstructure:"log_position"`

	// Pprof specifies if pprof endpoints are enabled.
	Pprof bool `mapstructure:"pprof"`

	// DisableDebug specifies if debug mode is disabled.
	DisableDebug bool `mapstructure:"-"`
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
	InlayConf   []interface{}
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
	return ztype.New(c.Core().Get(key))
}

// Unmarshal unmarshals the value associated with the given key in the Conf struct.
//
// It takes a string key and a pointer to an interface{} as its parameters.
// The function returns an error.
func (c *Conf) Unmarshal(key string, rawVal interface{}) error {
	return c.Core().UnmarshalKey(key, &rawVal)
}

// NewConf creates a new Conf object with the given options.
//
// opt: The optional configuration options.
//
//	These options are functions that modify the Conf object.
//	They can be used to customize the behavior of the Conf object.
//	The functions should accept a pointer to a gconf.Option object.
//	Example:
//	func(o *gconf.Option) {
//	    o.EnvPrefix = AppName
//	    o.AutoCreate = true
//	    o.PrimaryAliss = "dev"
//	}
func NewConf(opt ...func(o *gconf.Option)) func() *Conf {
	cfg := gconf.New(fileName, func(o *gconf.Option) {
		o.EnvPrefix = AppName
		o.AutoCreate = true
		o.PrimaryAliss = "dev"
		for _, f := range opt {
			f(o)
		}
	})

	return func() *Conf {
		c := &Conf{cfg: cfg}
		setConf(c, DefaultConf, false)
		common.Fatal(cfg.Read())
		setConf(c, InlayConf, true)
		common.Fatal(cfg.Unmarshal(&c))

		return c
	}
}

func (c *Conf) Core() *viper.Viper {
	return c.cfg.Core
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

func setConf(conf *Conf, value []interface{}, replace bool) {
	disableDebug := false
	set := conf.cfg.SetDefault
	if replace {
		set = conf.cfg.Set
	}
	toMap := func(isPrt bool, t reflect.Type, v reflect.Value) map[string]interface{} {
		m := make(map[string]interface{})

		for i := 0; i < t.NumField(); i++ {
			value, field := v.Field(i), t.Field(i)
			if value.IsZero() || !zstring.IsUcfirst(field.Name) {
				continue
			}

			m[field.Name] = v.Field(i).Interface()
		}
		return m
	}

	for _, c := range value {
		t := reflect.TypeOf(c)
		v := reflect.ValueOf(c)
		isPrt := t.Kind() == reflect.Ptr
		if isPrt {
			t = t.Elem()
			v = v.Elem()
		}

		name := getConfName(v)
		if t.Kind() != reflect.Struct {
			if t.Kind() == reflect.Slice {
				maps := make([]map[string]interface{}, 0)
				for i := 0; i < v.Len(); i++ {
					maps = append(maps, toMap(isPrt, t.Elem(), v.Index(i)))
				}
				set(name, maps)
			}
			continue
		}
		m := toMap(isPrt, t, v)
		if name == "base" {
			disableDebug = ztype.ToBool(m["DisableDebug"])
		}
		set(name, m)
	}

	if disableDebug {
		conf.Base.Debug = false
	}
}

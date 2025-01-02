package service

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sohaha/zlsgo/ztime"
	"github.com/zlsgo/app_core/common"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/ztype"
	gconf "github.com/zlsgo/conf"
)

type BaseConf struct {
	// LogDir specifies the directory for log files.
	LogDir string `z:"log_dir,omitempty"`

	// LogFile log file name
	LogFile string `z:"log_file,omitempty"`

	// LogShowDate specifies if log date should be included in logs.
	LogShowDate bool `z:"log_show_date,omitempty"`

	// Port specifies the port number for the server.
	Port string `z:"port,omitempty"`

	// CertFile CertFile
	CertFile string `z:"cert_file,omitempty"`

	// KeyFile KeyFile
	KeyFile string `z:"key_file,omitempty"`

	// HTTPAddr HTTPAddr
	HTTPAddr string `z:"http_addr,omitempty"`

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

	// HotReload specifies if hot reload is enabled.
	HotReload bool `z:"hot_reload,omitempty"`

	// DisableDebug specifies if debug mode is disabled.
	DisableDebug bool `z:"-"`
}

func init() {
	// Set the name of the configuration file.
	if ConfFileName == "" {
		ConfFileName = os.Args[0]
		ConfFileName = filepath.Base(ConfFileName)
		ConfFileName = strings.TrimSuffix(ConfFileName, filepath.Ext(ConfFileName))
	}
}

// ConfKey is used to get the configuration key.
func (BaseConf) ConfKey() string {
	return "base"
}

// DisableWrite disables writing to the configuration.
func (BaseConf) DisableWrite() bool {
	return true
}

func (b BaseConf) fix(c *gconf.Confhub, bc BaseConf) BaseConf {
	var m ztype.Map
	_ = c.UnmarshalKey(b.ConfKey(), &m)

	if !m.Has("zone") {
		bc.Zone = baseConf.Zone
	}

	if !m.Has("port") {
		bc.Port = baseConf.Port
	}

	if !m.Has("HotReload") {
		bc.HotReload = baseConf.HotReload
	}

	return bc
}

func (b BaseConf) Reload(r *Web, conf *Conf) {
	if !baseConf.HotReload {
		return
	}

	var nb BaseConf
	err := conf.Unmarshal(nb.ConfKey(), &nb)
	nb = nb.fix(conf.cfg, nb)
	if err != nil || reflect.DeepEqual(baseConf, nb) {
		return
	}

	var port int
	addr := strings.SplitN(nb.Port, ":", 2)
	if len(addr) != 2 {
		port = ztype.ToInt(port)
	} else {
		port = ztype.ToInt(addr[1])
	}

	port, _ = znet.Port(port, false)
	if port == 0 {
		zlog.Errorf("port is not valid:%s", nb.Port)
		return
	}

	_ = r.Restart()
}

var (
	// ConfFileName is the name of the configuration file.
	ConfFileName = ""

	// LogPrefix is the prefix for log messages.
	LogPrefix = ""

	// AppName is the name of the application.
	AppName = "ZlsAPP"

	// debug determines whether debug mode is enabled.
	debug = false
)

var (
	DefaultConf []interface{}
	baseConf    = BaseConf{
		Debug:     debug,
		Zone:      8,
		Port:      "3788",
		HotReload: true,
	}
)

// Conf represents the configuration struct.
type Conf struct {
	cfg *gconf.Confhub // cfg is used to manage the configuration settings.

	Base          BaseConf      // Base represents the base configuration settings.
	autoUnmarshal func()        `z:"-"`
	reloads       []interface{} `z:"-"`
}

// Get retrieves the value associated with the given key from the Conf object.
func (c *Conf) Get(key string) ztype.Type {
	return c.cfg.GetAll().Get(key)
}

// Set updates the value of a configuration key.
func (c *Conf) Set(key string, value interface{}) {
	c.cfg.Set(key, value)
}

// Unmarshal unmarshals the value associated with the given key in the Conf struct.
// It takes a string key and a pointer to an interface{} as its parameters.
func (c *Conf) Unmarshal(key string, rawVal interface{}) error {
	return c.cfg.UnmarshalKey(key, &rawVal)
}

// NewConf creates a new Conf object with the given options.
func NewConf(opt ...func(o gconf.Options) gconf.Options) func(di zdi.Injector) *Conf {
	cfg := gconf.New(ConfFileName, func(o gconf.Options) gconf.Options {
		o.EnvPrefix = AppName
		o.AutoCreate = true
		o.PrimaryAliss = "dev"
		for i := range opt {
			o = opt[i](o)
		}
		return o
	})

	return func(di zdi.Injector) *Conf {
		c := &Conf{cfg: cfg}

		delay, autoUnmarshal := setConf(c, DefaultConf)

		common.Fatal(cfg.Read())
		delay()
		autoUnmarshal()

		common.Fatal(cfg.Unmarshal(&c))

		// Because the basic configuration is not a pointer type, we need to reassign it here.
		baseConf = baseConf.fix(cfg, c.Base)
		c.Base = baseConf

		c.autoUnmarshal = autoUnmarshal

		ztime.SetTimeZone(int(c.Base.Zone))

		return c
	}
}

type DefaultConfValue interface {
	ConfKey() string
	DisableWrite() bool
}

// RegisterDefaultConf registers a default configuration value.
func RegisterDefaultConf(conf DefaultConfValue) {
	DefaultConf = append(DefaultConf, conf)
}

func (c *Conf) Write() error {
	return c.cfg.Write()
}

func getConfName(t reflect.Value) (key string, isVar bool) {
	getConfKey := t.MethodByName("ConfKey")
	if getConfKey.IsValid() {
		g, ok := getConfKey.Interface().(func() string)
		if ok {
			key = g()
			isVar = true
		}
	}

	if key == "" {
		if t.Kind() == reflect.Slice {
			key = t.Type().Elem().Name()
		} else {
			key = t.Type().Name()
		}
	}

	return
}

func setConf(conf *Conf, value []interface{}) (func(), func()) {
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
		i := i
		v := reflect.ValueOf(value[i])
		isPtr := v.Kind() == reflect.Ptr
		name, _ := getConfName(v)

		d := v.MethodByName("DisableWrite")
		disableWrite := false
		if d.IsValid() {
			if f, ok := d.Interface().(func() bool); ok {
				disableWrite = f()
			}
		}

		r := v.MethodByName("Reload")
		if r.IsValid() && r.Kind() == reflect.Func {
			// TODO: 考虑使用 reflect.DeepEqual 来判断是否需要通知
			conf.reloads = append(conf.reloads, r.Interface())
		}

		set := setConf(disableWrite)
		typ := reflect.Indirect(v).Type()

		switch typ.Kind() {
		case reflect.Struct:
			m := ztype.ToMap(value[i])
			if name == "base" {
				disableDebug = m.Get("DisableDebug").Bool()
			}
			set(name, m)
		case reflect.Slice:
			switch typ.Elem().Kind() {
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
			set(name, value[i])
		}

		if isPtr {
			autoUnmarshal = append(autoUnmarshal, func() {
				_ = conf.cfg.UnmarshalKey(name, value[i], true)
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
		}, func() {
			for _, f := range autoUnmarshal {
				f()
			}
		}
}

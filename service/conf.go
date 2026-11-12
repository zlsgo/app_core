package service

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

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

	// LogLevel log level
	LogLevel string `z:"log_level,omitempty"`

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

	// LogMaxAge log max age
	LogMaxAge int `z:"log_max_age,omitempty"`
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

	if !m.Has("hot_reload") {
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

	port := parsePort(nb.Port)
	port, _ = znet.Port(port, false)
	if port == 0 {
		zlog.Errorf("port is not valid:%s", nb.Port)
		return
	}

	_ = r.Restart()
}

// parsePort extracts the port from values such as "8080", ":8080" or
// "127.0.0.1:8080". It returns 0 when the input is not a port.
func parsePort(addr string) int {
	addr = strings.TrimSpace(addr)
	if i := strings.LastIndex(addr, ":"); i >= 0 {
		addr = addr[i+1:]
	}
	return ztype.ToInt(addr)
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

	Base BaseConf // Base represents the base configuration settings.

	mu         sync.Mutex    // mu guards applied, as module construction and conf reads may happen in different places.
	applied    int           // applied is the number of DefaultConf entries already applied.
	unmarshals []func()      `z:"-"` // unmarshals holds closures that unmarshal file values back into default pointers; rerun on hot reload.
	reloads    []interface{} `z:"-"`
}

// Get retrieves the value associated with the given key from the Conf object.
func (c *Conf) Get(key string) ztype.Type {
	return c.cfg.GetAll().Get(key)
}

// GetAll retrieves all configuration values as a map.
func (c *Conf) GetAll() ztype.Map {
	return c.cfg.GetAll()
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
		o.EnvPrefix = strings.ToUpper(AppName)
		o.AutoCreate = true
		o.PrimaryAliss = "dev"
		o.EnvKeyReplacer = "_"

		for i := range opt {
			o = opt[i](o)
		}
		return o
	})

	return func(di zdi.Injector) *Conf {
		c := &Conf{cfg: cfg}

		delay := setConf(c, DefaultConf)
		c.applied = len(DefaultConf)

		common.Fatal(cfg.Read())
		delay()
		c.refresh()

		common.Fatal(cfg.Unmarshal(&c))

		// Because the basic configuration is not a pointer type, we need to reassign it here.
		baseConf = baseConf.fix(cfg, c.Base)
		c.Base = baseConf

		ztime.SetTimeZone(int(c.Base.Zone))

		return c
	}
}

// ApplyDefaults applies default configurations registered after NewConf ran.
//
// NewConf snapshots DefaultConf at construction time, so entries appended
// afterwards are never seen. Modules typically register their defaults in
// New(), and if *service.Conf is constructed first (e.g. with a
// RegModule(cfg *service.Conf) signature), those module defaults are lost.
// InitModule calls this before the module's Load to land late-registered
// defaults.
func (c *Conf) ApplyDefaults() {
	c.mu.Lock()
	if c.applied >= len(DefaultConf) {
		c.mu.Unlock()
		return
	}
	pending := make([]interface{}, len(DefaultConf)-c.applied)
	copy(pending, DefaultConf[c.applied:])
	c.applied = len(DefaultConf)
	c.mu.Unlock()

	for i := range pending {
		setConf(c, pending[i:i+1])()
	}

	// SetDefault does not refresh the conf cache; rebuild it once so subsequent
	// non-forced Get/Unmarshal calls see the new defaults.
	c.cfg.GetAll(true)
	c.refresh()
}

// refresh re-unmarshals the config file values back into each module's Options pointer.
func (c *Conf) refresh() {
	for _, f := range c.unmarshals {
		f()
	}
}

type DefaultConfValue interface {
	ConfKey() string
	DisableWrite() bool

	// Preset sets the default values for the configuration. Non-compulsory.
	// Preset()

	// Reload is called when the configuration is reloaded. Non-compulsory.
	// Reload()
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

// setConf writes a batch of default configurations into the conf hub and
// returns the deferred actions that must run after Read.
//
// Pointer-typed configs also register an unmarshal closure: after Read, file
// values are written back into the original struct (only keys present in the
// file are overwritten), so code-level defaults set in New(opt) are not wiped.
func setConf(conf *Conf, values []interface{}) (delay func()) {
	confs, disableDebug := ztype.Map{}, false
	set := func(disableWrite bool) func(key string, value interface{}) {
		if !disableWrite {
			return conf.cfg.SetDefault
		}
		return func(key string, value interface{}) {
			confs[key] = value
		}
	}

	for idx := range values {
		i := idx
		v := reflect.ValueOf(values[i])
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
			// TODO: consider using reflect.DeepEqual to decide whether to notify
			conf.reloads = append(conf.reloads, r.Interface())
		}

		p := v.MethodByName("Preset")
		if p.IsValid() && p.Kind() == reflect.Func && p.Type() == reflect.TypeOf(func() {}) {
			p.Call(nil)
		}

		put := set(disableWrite)
		typ := reflect.Indirect(v).Type()

		switch typ.Kind() {
		case reflect.Struct:
			m := ztype.ToMap(values[i])
			if name == "base" {
				disableDebug = m.Get("DisableDebug").Bool()
				put(name, map[string]interface{}(m))
			} else {
				put(name, m)
			}
		case reflect.Slice:
			switch typ.Elem().Kind() {
			case reflect.Struct:
				m := make([]map[string]interface{}, v.Len())
				for i := 0; i < v.Len(); i++ {
					m[i] = ztype.ToMap(v.Index(i).Interface())
				}
				put(name, m)
			default:
				put(name, v)
			}
		default:
			put(name, values[i])
		}

		if isPtr {
			conf.unmarshals = append(conf.unmarshals, func() {
				_ = conf.cfg.UnmarshalKey(name, values[i], true)
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
	}
}

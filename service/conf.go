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
	LogDir      string
	Port        string
	PprofToken  string
	Zone        int8 `mapstructure:"Zone"`
	Debug       bool
	LogPosition bool
	Pprof       bool
	Statsviz    bool
}

func (BaseConf) ConfKey() string {
	return "base"
}

var (
	// fileName 配置文件名
	fileName = "conf"
	// LogPrefix 日志前缀
	LogPrefix = ""
	// AppName 项目名称
	AppName = "ZlsAPP"
	// debug 设置生成配置时，程序默认运行模式
	debug = true
)

var (
	DefaultConf []interface{}
)

// Conf 配置项
type Conf struct {
	cfg *gconf.Confhub

	Base BaseConf
}

func (c *Conf) Get(key string) ztype.Type {
	return ztype.New(c.Core().Get(key))
}

func (c *Conf) Unmarshal(key string, rawVal interface{}) error {
	return c.Core().UnmarshalKey(key, &rawVal)
}

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

		for _, c := range DefaultConf {
			t := reflect.TypeOf(c)
			v := reflect.ValueOf(c)
			isPrt := t.Kind() == reflect.Ptr
			if isPrt {
				t = t.Elem()
				v = v.Elem()
			}

			if t.Kind() != reflect.Struct {
				if t.Kind() == reflect.Slice {
					maps := make([]map[string]interface{}, 0)
					for i := 0; i < v.Len(); i++ {
						maps = append(maps, toMap(isPrt, t.Elem(), v.Index(i)))
					}
					cfg.SetDefault(getConfName(v), maps)
				}
				continue
			}
			cfg.SetDefault(getConfName(v), toMap(isPrt, t, v))
		}

		common.Fatal(cfg.Read())
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

package service

import (
	"reflect"

	"github.com/zlsgo/app_core/utils"

	"github.com/sohaha/zlsgo/ztype"
	"github.com/spf13/viper"
	gconf "github.com/zlsgo/conf"
)

type base struct {
	Zone        int8   `mapstructure:"Zone"` // 时区区域
	Debug       bool   // 开启全局调试模式
	LogDir      string // 日志目录
	LogPosition bool   // 调试下打印日志显示输出位置
	Port        string // 项目端口
	Pprof       bool   // 开启 pprof
	Statsviz    bool   // 开启 statsviz
	PprofToken  string // pprof Token
}

const (
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

func init() {
	DefaultConf = append(DefaultConf, base{
		Debug: debug,
		Zone:  8,
		Port:  "3788",
	})
}

// Conf 配置项
type Conf struct {
	cfg *gconf.Confhub

	Base base
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
				if value.IsZero() {
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

		utils.Fatal(cfg.Read())
		utils.Fatal(cfg.Unmarshal(&c))

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

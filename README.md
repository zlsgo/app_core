# app_core

基于 [zlsgo](https://github.com/sohaha/zlsgo) 的应用内核，封装配置、日志、依赖注入、Web 路由、定时任务和模块生命周期。它本身是 Go 库，不含 `main` 函数。

> 新建业务项目可以直接使用 [zlsgo-app](https://github.com/sohaha/zlsgo-app) 作为项目基础脚手架，它自带 `main.go`、示例模块、控制器、错误处理和构建脚本。更完整的说明见 [zlsgo 使用文档](https://docs.73zls.com/zlsgo/#/89a1532c-15bd-495e-a427-30b3cff6e061)，本文只介绍 app_core 的简单用法。

## 安装

Go 1.25 或更高版本：

```bash
go get github.com/zlsgo/app_core
```

## 快速开始

```go
package main

import (
	"context"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/znet"
	"github.com/zlsgo/app_core/service"
)

func main() {
	di := zdi.New()
	di.Map(di, zdi.WithInterface((*zdi.Injector)(nil)))
	di.Map(context.Background())

	// 配置文件名默认取可执行文件名，如 tmpApp.toml
	di.Provide(service.NewConf())

	// 可通过代码覆盖基础配置，也可写在配置文件中
	di.Provide(service.NewApp(func(o service.BaseConf) service.BaseConf {
		o.Port = "8181"
		return o
	}))

	di.Provide(service.NewWeb())

	// 注册全局中间件、控制器和模块
	di.Provide(func() []znet.Handler { return nil })
	di.Provide(func() *[]service.Controller {
		return &[]service.Controller{NewIndex()}
	})
	di.Provide(func() []service.Module { return nil })

	// 加载模块生命周期：Load -> Start -> Done
	if err := di.InvokeWithErrorOnly(service.InitModule); err != nil {
		panic(err)
	}

	// 启动全局定时任务和 Web 服务，阻塞直到服务退出
	if err := di.InvokeWithErrorOnly(service.InitTask); err != nil {
		panic(err)
	}
	if err := di.InvokeWithErrorOnly(service.RunWeb); err != nil {
		panic(err)
	}
}

// 控制器实现 Init(*znet.Engine) error；方法名由 znet 绑定为 HTTP 路由
type Index struct {
	service.App // 自动注入 DI、Conf、Log
	Path string // 指定路由前缀；留空则按控制器包/类型名推导（以 Index 结尾时映射到根路径）
}

func NewIndex() *Index { return &Index{Path: "api"} }

func (h *Index) Init(r *znet.Engine) error { return nil }

func (h *Index) GET(c *znet.Context) string {
	return "ok"
}
```

编译运行后访问 `http://127.0.0.1:8181/api`：

```bash
go build -o tmpApp && ./tmpApp
```

配置文件默认名为 `<可执行文件名>.toml`，路径以 zlsgo 的 `zfile.ProjectPath` 为基准：通常是可执行文件所在目录；当可执行文件位于临时目录（例如 `go run`）时回退到启动时的当前工作目录。可手动创建；`<名称>-dev.toml` 会覆盖主配置文件，适合作为本地开发配置。示例：

```toml
[base]
port = "8181"
debug = true
log_level = "debug"
log_dir = "./logs"
zone = 8
```

默认端口 `3788`、时区 `8`。常用项还有 `log_file`、`log_show_date`、`log_position`、`log_max_age`、`cert_file` / `key_file` / `http_addr`（HTTPS）、`pprof` / `pprof_token`。`hot_reload` 控制基础配置变更时是否重启 Web，默认开启，可写在代码或配置文件中。配置格式详见 [zlsgo/conf](https://github.com/zlsgo/conf)，也可调用 `Conf.Write()` 写出配置。

环境变量覆盖由 `NewConf` 的 options 开启。在 import 区加入 `gconf "github.com/zlsgo/conf"`，把快速开始里的 `di.Provide(service.NewConf())` 换成：

```go
di.Provide(service.NewConf(func(o gconf.Options) gconf.Options {
	o.AutomaticEnv = true
	return o
}))
```

默认前缀为 `ZLSAPP`，可在调用 `NewConf` 前修改 `service.AppName`。分层 key 用点号，例如：

```bash
env 'ZLSAPP_BASE.PORT=8080' ./tmpApp
```

这会覆盖 `base.port`。当前未设置点号转下划线替换器，`ZLSAPP_BASE_PORT` 不生效。`BaseConf.AutomaticEnv` 和配置文件里的 `automatic_env` 只是配置值，不负责开启环境变量覆盖。

注意：环境变量只对已在配置里出现、或代码默认值非零（如 `port`）的键生效；没有先例的键（如 `debug`、`log_level` 等默认空值）即使设置了环境变量也不会进入 Conf。

## 模块使用

模块由 `service.InitModule` 统一加载，先注册到 DI：

```go
di.Provide(func() []service.Module {
	return []service.Module{newDemo()}
})
```

生命周期按模块名排序执行 `Load` → 全部模块 `Start` → 注册路由/定时任务并调用 `Done` → `RunWeb` 启动服务 → 退出时调用可选的 `Stop`。`Load` 的返回值是函数时会被 `di.Provide`，是其他值则 `di.Map`，可用于模块间共享服务；模块内的 `App`、`DI`、`Conf`、`Log` 字段会自动注入。

### 1. 回调式：ModuleLifeCycle

嵌入 `service.ModuleLifeCycle` 即可用回调定义生命周期，不必手写接口方法：

```go
type Demo struct {
	service.ModuleLifeCycle
}

func (d *Demo) Name() string { return "Demo" } // 可选，默认取包名

func newDemo() *Demo {
	return &Demo{
		ModuleLifeCycle: service.ModuleLifeCycle{
			OnLoad: func(di zdi.Invoker) (any, error) {
				// 配置解析完成后执行；返回值注入 DI 供其他模块使用
				return func() string { return "shared" }, nil
			},
			OnStart: func(di zdi.Invoker) error { return nil },
			OnDone:  func(di zdi.Invoker) error { return nil },
			OnStop:  func(di zdi.Invoker) error { return nil },
			Service: &service.ModuleService{
				Controllers: []service.Controller{},
				Tasks: []service.Task{
					{Name: "demo", Cron: "1 * * * * * *", Run: func() {}},
				},
			},
		},
	}
}
```

模块配置用 `ConfKey() string` 定义 key，追加到 `service.DefaultConf`；若用 `service.RegisterDefaultConf()` 注册还需实现 `DisableWrite() bool`。可选实现 `Preset()`、`Reload()`，配置文件变更会重新反序列化并触发 `Reload`。

### 2. modules/single：泛型单模块

适合在 `Done` 阶段产出一个实例，用 `Instance()` 取回：

```go
import "github.com/zlsgo/app_core/modules/single"

worker := single.New(single.Lifecycle[string]{
	Name: "Worker",
	Load: func(di zdi.Invoker) (any, error) { return nil, nil },
	Done: func(di zdi.Invoker) (string, error) { return "ready", nil },
})

_ = worker.Instance() // Done 执行后可用
```

`single.Module` 已实现 `service.Module`，直接放进 `[]service.Module`；它不提供控制器和定时任务。

### 3. modules/multiple：聚合多个模块

把多个模块打包成一个 `service.Module`，汇总子模块的 `Tasks`、`Controller`，并在 `Load` 中依次调用子模块 `Load`：

```go
import "github.com/zlsgo/app_core/modules/multiple"

bundle := multiple.New(map[string]service.Module{
	"user":  newDemo(),
	"order": newDemo(),
})

bundle.Add("pay", newDemo()) // 追加
m, ok := bundle.Get("user")  // 按名字取回
_, _ = m, ok
```

`multiple.New` 使用包级存储，同一进程只应初始化一次；同名模块请用不同 key 包装，避免被去重。子模块按注册 key 的字典序执行 `Load` → `Start` → `Done`，`Tasks`、`Controller` 和 `Name` 也按该顺序汇总，便于复现启动顺序。

### 4. 直接实现 service.Module

需要完全控制时实现完整接口：

```go
type Custom struct {
	service.App // 自动注入 DI、Conf、Log
}

func (c *Custom) Name() string                     { return "Custom" }
func (c *Custom) Tasks() []service.Task            { return nil }
func (c *Custom) Controller() []service.Controller { return nil }
func (c *Custom) Load(di zdi.Invoker) (any, error)  { return nil, nil }
func (c *Custom) Start(di zdi.Invoker) error        { return nil }
func (c *Custom) Done(di zdi.Invoker) error         { return nil }
```

可运行的业务示例见 [zlsgo-app 的 module/example](https://github.com/sohaha/zlsgo-app/tree/module-web/module/example)。

## 其他入口

- 全局对象：`app_core.Conf()`、`app_core.Log()`、`app_core.DI()` 返回 `error`，未初始化或对应字段为空时返回 `app_core.ErrNotInitialized`；`MustConf()` 等旧接口会 panic，已不推荐。
- 控制器：实现 `Init(*znet.Engine) error`，`Path` 字段指定路由前缀，也可在 `Init` 内用 `r.GET` 手动注册路由。
- 定时任务：`service.Task{Name, Cron, Run}`，Cron 表达式见 [zlsgo](https://github.com/sohaha/zlsgo)。
- 公共工具：`common.VarPages`、`common.VarUID`、`common.Upload`、`common.Fatal`。

### 文件上传

`common.Upload(c, subDirName, opts...)` 把表单文件保存到 `common.Define.UploadLocalDir`（默认 `./static`，同样相对 `zfile.ProjectPath`）：

```go
files, err := common.Upload(c, "avatar", func(o *common.UploadOption) {
	o.Key = "file"                     // 表单字段名，默认 file
	o.Dir = "users"                    // 追加到 subDirName 之后；以 "/" 开头时忽略 subDirName
	o.MimeType = []string{"image/png"} // 允许的 MIME，留空不限制
	o.MaxSize = 2 * 1024 * 1024        // 单文件大小上限，默认 2MiB
})
```

`subDirName` 和 `Dir` 按相对目录处理：开头的分隔符会被忽略，包含 `..` 的值直接返回错误，避免把文件写到上传根目录之外。使用编译后的二进制时，返回的 `Path` 形如 `/static/xxx`，可直接作为静态资源路径；`go run` 等可执行文件位于临时目录的场景会返回绝对路径。

## 相关项目

- [zlsgo-app](https://github.com/sohaha/zlsgo-app)：项目基础脚手架（默认分支 `module-web`）
- [zlsgo](https://github.com/sohaha/zlsgo)：底层工具库
- [zlsgo/conf](https://github.com/zlsgo/conf)：配置文件读写与热重载
- [使用文档](https://docs.73zls.com/zlsgo/#/89a1532c-15bd-495e-a427-30b3cff6e061)

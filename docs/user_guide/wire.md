# wire依赖注入

框架在biz层和handler层都有 `wire.go` ，它会在执行 `make generate` 成功后，覆写同目录下的 `wire_gen.go` ，生成完成依赖注入代码的接口函数供调用。

wire 是google开源的一个依赖注入组件，参考 [google/wire](https://github.com/google/wire)

它的官方文档见：[wire/_tutorial/README.md](https://github.com/google/wire/blob/main/_tutorial/README.md)

简单来说就是，在提供接口的地方定义 ProviderSet，如： `internal/biz/base.go` 里的 `ProviderSet` ；然后在用到接口的地方，引用ProviderSet，如： `internal/biz/wire.go` 里的 `func WireBizService` 。
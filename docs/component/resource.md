# 资源(resource)

每个不同的后端服务都是一份资源，如：DB，Redis，OSS，其它微服务等。另外，每个第三方组件，也是一份资源，如：jwt等。<font color=#24aae6>**这里的核心观点是，这些附加资源和它们所附属的部署保持松耦合。**</font>（参考：[The Twelve-Factor App （简体中文） (12factor.net)](https://12factor.net/zh_cn/backing-services)）

那么，这些资源的实例和配置，都应该在 resource 组件（ `internal/common/resource` ）里进行初始化和管理。

请注意，每份资源的Config的定义，都应该在资源内部，resource组件只是引用它。请不要将不同资源的Config都放到同一处地方进行定义，这会破坏资源的松耦合。
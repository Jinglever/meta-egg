# 配置管理(config和constraint)

框架把“配置”划分了三大类别：基础设施类，业务约束类，资源类。

对照 `configs/conf.yml` 理解这几大分类：

- 基础设施类：如 `log` ， `http_server` ， `grpc_server` ， `monitor_server` 都属于基础设施类。它们可以认为是业务无关的配置，更偏向于系统的基础支撑。
- 业务约束类：在 `constraint` 下面可以放置与业务相关的配置，如：cron定时任务规则，是否开启上报等。
- 资源类：在 `resource` 下放置资源类的配置。那什么是资源呢？每个不同的后端服务都是一份资源，如：DB，Redis，OSS，其它微服务等。另外，每个第三方组件，也是一份资源，如：jwt等。<font color=#24aae6>**这里的核心观点是，这些附加资源和它们所附属的部署保持松耦合。**</font>（参考：[The Twelve-Factor App （简体中文） (12factor.net)](https://12factor.net/zh_cn/backing-services)）
# swagger文档

适用于http服务。

在 `internal/server/http/router.go` 里定义swagger文档的 title 和 version 等。

在handler层各接口函数的上方，按照swagger约定的格式，编写注释，它会在执行 `make swag` 命令的时候，在工程的 `docs` 目录下生成相应的接口swagger文档。

参考示例：

```go
//	@Id			GetUserDetail
//	@Tags		用户
//	@Summary	获取用户详情
//	@Description
//	@Accept		json
//	@Produce	json
//	@Param		Authorization	header		string	true	"Bearer <jwt-token>"
//	@Param		id				path		int		true	"用户ID"
//	@Success	200				{object}	RspData{data=UserDetail}
//	@Failure	400				{object}	RspBase
//	@Router		/v1/users/{id} [get]
```
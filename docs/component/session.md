# 会话管理和校验(ME和jwt)

对于会话的身份校验，框架默认提供了jwt，参见 `internal/server/http/middleware.go` 里的 `authHandler` 和 `internal/server/grpc/middleware.go` 里的 `authInterceptor` 。

对于http请求，要在header中添加名为 `Authorization` 的key，值的格式为 `Bearer <jwt-token>` 。

对于grpc请求，要在 metadata 中添加名为 `authorization` 的key，值的格式为 `Bearer <jwt-token>` 。

以上的key-value设置，均符合最常见的用法。并且，它们在postman中被原生支持。

 ![postman-http](/images/postman_http.png)

 ![postman-grpc](/images/postman_grpc.png)

 另外，在编码仅jwt token的 Payload 里的内容，应该是 `contexts.ME` 结构体做json序列化之后的字符串。
# 对象转换规范(VO <-> BO <-> Model)

框架给出的代码示例里，分别在handler、biz定义了VO（view object）和BO（business object），它们存在的意义是作为胶水层将不同层之间的数据结构解耦，提高代码的可扩展性。它们的代码示例分别如下：

- VO： `internal/handler/http/user.go` 的 `UserDetail` 、 `UserListInfo` 等
- BO： `internal/biz/user.go` 的 `UserBO` 、 `UserListBO` 等
- Model： `gen/model/user.go` 的 `User`

VO跟BO之间的转换，在handler层进行；BO跟Model之间的转换，在biz层进行。
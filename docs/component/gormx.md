# gorm再封装(gormx)

gorm本身是一个非常优秀并且友好的 ORM 组件。为什么要在它上面再封装一层呢？因为gorm太过强大和灵活，而我们在实际工程中，往往只需要用到它的某种配置搭配，以及大多情况下只会用到它的若干接口，并且是有限的几种用法。所以，封装一层，把团队习惯的配置搭配和接口用法包装出来，会使得写业务代码时更加轻松，更加快速。另外，还可以增加新的特性，从而更好的支持框架的设计。

以下分几点来看：

#### 连接数据库
连接数据库需要在工程里暴露若干配置项，它是gorm的配置项的子集，比如：数据库类型，DSN，日志level等。另外有些配置项，可以写死在代码里作为默认配置，比如 SkipDefaultTransaction=true，PrepareStmt=false，TranslateError=true等。    
参见 `pkg/gormx/connect.go` 。
    
#### 支持事务嵌套
gorm原生的数据库事务的写法有两种，而要支持事务嵌套的话，必须使用闭包的写法。在这个写法里，作为传递介质的是 `*gorm.DB` 实例。    
然后，框架期望借助golang最常见的context来传递，所以在gormx里封装出来一个新的 DB 对象，支持新的 `Transaction` 接口，通过 context 来传递数据库句柄。
参见 `pkg/gormx/db.go` 。
    
#### 数据库查询接口
gorm的数据库查询接口是链式调用的模式，但我们在写业务逻辑的时候，往往通过slice类型数据结构更便于操作查询条件的组合。所以，gormx封装成了组合option的调用模式。
参见 `pkg/gormx/option.go` ，也可以阅读 `gen/repo` 里的代码，了解option的用法。
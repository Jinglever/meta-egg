> before 2024-01-01
- [x] add GetIDBySemantic() into gen/repo for meta table

> before 2024-01-01
- [x] fix meta sql generation bug
- [x] ready for open source

> before 2023-12-30
- [x] Add SQL statements to disable foreign key checks and set target database schema for meta-egg.sql

> before 2023-12-28
- [x]default fitting error context into cerror for return, config to decide if return error detail from handler
- [x]alter makefile and dockerfile to support gitlab CI/CD better

> before 2023-12-22
- [x]table in demo with name "gender", change name to "mt_gender"
- [x]change the name "jwt" to "access_token"
- [x]give choice for use while execute "meta-egg new" that if need auth (such as access_token) or not
- [x]in http router, change variable name "authGroup" to "apiGroup"
- [x]change http handler path prefix from "/v1" to "/api/v1"
- [x]in proto, create a proto file to define message for every entity set `handler=true` in manifest.xml
- [x]support to generate handlers for meta table

> before 2023-12-01
- [x] Update config file paths and database connection

> before 2023-11-21
- [x] Fix typos and rename functions in error and logging modules
- [x] Fix time parsing bug in handler_generator

> before 2023-11-08
- [x] 给custom template增加占位符%%PROJECT-NAME-DIR%%、%%PROJECT-NAME-STRUCT%%
- [x] update generate command
- [x] add confirmation for project root directory when using `update` command

> before 2023-11-08
- [x] 给custom template增加占位符%%PROJECT-NAME-PKG%%

> before 2023-11-07
- [x] 新增的文件仍受ignore_files里的精准匹配的影响，但不受ignore_files里的通配符匹配的影响
- [x] gen里的文件同样受ignore_files里的精准匹配的影响，但不受ignore_files里的通配符匹配的影响
- [x] 开放gorm的max_open、max_idle、max_lifetime、max_idle_time的配置
- [x] 对于required为true的字符串类型字段，如果设置了default_empty为true，则在http handler的create请求里，tag设为omitempty
- [x] 对于required为true的数值字段，如果设置了insert_default为0，则在http handler的create请求里，tag设为omitempty
- [x] 对于对应到code里是bool类型的字段，http handler的create请求里，tag避免出现required
- [x] 修复grpc handler的list接口，当不存在任何filter字段时未删除占位符的问题

> before 2023-10-18
- [x] 修复[]byte类型在filter和update处误用指针符号的问题

> before 2023-10-17
- [x] 支持设置CHAR/VARCHAR/TEXT类型字段的默认值为空字符串
- [x] 修复处理字段默认值的bug

> before 2023-10-15
- [x] 修复bug: proto的pagination字段应该小写开头
- [x] 修复bug: 生成的grpc的handler里option.Pagination的初始化问题
- [x] 支持新数据库字段类型：
    - for mysql/tidb: TINYBLOB|BLOB|MEDIUMBLOB|LONGBLOB
    - for postgres: BYTEA

> before 2023-09-12
- [x] 重构代码结构:
    - 将类sql查询的代码下沉到repo层
    - 在biz层增加BO，Model和BO之间的转换
    - 在handler层增加VO，BO和VO之间的转换
    - 调整monitor server的路由，使用pprof自带的
- [x] 给GRPC的handler增加create、list、delete接口
- [x] 调整demo的数据库结构，更全面地展示生成代码示例
- [x] 对于不开启handler的实体，不生成它的biz层代码，使biz层更加清爽
- [x] 给new命令增加对扩展代码模板的支持

> before 2023-09-06
- [x]修复生成pg的删字段的增量sql的问题，用双引号包裹字段名
- [x]对pg的date字段类型也做应用层的时区修正

> before 2023-09-02
- [x] biz层判断如果存在unique的字段，就给create和update增加对duplicate error的处理
- [x] meta表的smt字段改成用varchar，增加专用字段desc，用于生成代码时的注释
- [x] 扩展模板的placeholder增加project_name和project_desc
- [x] 增加命令查询当前支持的扩展模板placeholder列表: `meta-egg help template`

> before 2023-08-29
- [x] 支持在外部扩展模板代码

> before 2023-08-27
- [x] 优化wire的用法，将handler和bizService的New函数放到各自的包里，便于复用

> before 2023-08-26
- [x] 支持在env.yml里指定ignore给定数据库表，应对实际工程中使用了会自动创建数据库表的包的情况
- [x] 将"*_gen.go"预置进env.yml的ignore_files里
- [x] 生成的schema和meta_data自动更新到工程的sql目录内

> before 2023-08-24
- [x] 增加pg特有的timetz和timestamptz类型的支持
- [x] 针对time和datetime类型，在选用postgres的情况下，利用gorm的hook，在gen的代码里增加AfterFind，转为local时区的time.Time
- [x] 修复按meta表数据生成GetSematicByID函数里的字符串带有空格的问题

> before 2023-08-19
- [x] gormx的connectDB的gorm.config增加TranslateError:true

> before 2023-08-19
- [x] 在make generate命令里加入更新gorm到最新版本的操作
- [x] 将meta_egg.dtd也加到update操作的更新对象里
- [x] --uncertain的时候，区分提醒新增、更新，并特别提醒base.go的更新，重新设计提示文本的颜色

> before 2023-08-11
- [x] env.yml里的可以配置数据库名，有的话覆盖xml里的那个

> before 2023-08-01
- [x] dao层增加count方法，增加group的option
- [x] 自动生成META表数据的SQL文件，并自动覆盖到sql目录下
- [x] don't name table to 'base'
- [x] fix bug inc sql try to delete index for foreign key
package template

var TplDbTypePostgres = `postgres`
var TplDbDsnPostgres = `host=<host> port=<port> user=<username> password=<pwd> dbname=<database> sslmode=disable TimeZone=Asia/Shanghai target_session_attrs=read-write`

var TplDbTypeMysql = `mysql`
var TplDbDsnMysql = `<username>:<pwd>@tcp(<host>:<port>)/<database>?charset=utf8mb4&timeout=5s&parseTime=True&loc=Asia%2FShanghai`

var TplConfYmlResourceDB = `  # 数据库配置
  db:
    # 数据库类型, 可选值: mysql | postgres
    type: %%TPL-DB-TYPE%%
    # 数据库host
    host: "<host>"
    # 数据库端口
    port: "<port>"
    # 数据库用户名
    user: "<username>"
    # 数据库密码
    password: "<password>"
    # 数据库名
    database: "<database>"
    # 时区
    timezone: "Asia/Shanghai"
    # 日志级别 info|warn|error|silent
    log_level: "warn"
    # 最大连接数 <=0表示无限制
    max_open: 0
    # 最大空闲连接数 <=0表示不保留空闲连接 默认值: 2
    # 如果max_idle>max_open且max_open>0则max_idle=max_open
    max_idle: 2
    # 最大连接存活时 <=0代表无限制 格式如: 1m 代表1分钟
    max_lifetime: 0
    # 最大空闲时间 <=0代表无限制 格式如: 1m 代表1分钟
    max_idle_time: 0
`

var TplConfYmlResourceAccessToken = `  # access_token JWT配置
  access_token:
    # 会话有效时长
    max_age: 1h
    # HS256相关
    hs256_secret_is_base64: false # 密钥是否为base64编码
    hs256_secret: "xxxxx" # 密钥
`

var TplConfYmlHttpServer = `http_server:
  # HTTP服务监听地址
  address: ":8080"
  # HTTP服务读取请求头超时时间
  read_header_timeout: 1s
  # http接口是否返回错误详情
  return_error_detail: false
%%TPL-CONF-YML-HTTP-SERVER-ACCESS-TOKEN%%
`

var TplConfYmlHTTPServerAccessToken = `  # 是否验证access token, 为false时, 会仅解析JWT, 不会验证JWT签名
  verify_access_token: true
`

var TplConfYmlGrpcServer = `grpc_server:
  # GRPC服务监听地址
  address: ":8081"
  # http接口是否返回错误详情
  return_error_detail: false
%%TPL-CONF-YML-GRPC-SERVER-ACCESS-TOKEN%%
`

var TplConfYmlGRPCServerAccessToken = `  # 是否验证access token, 为false时, 会仅解析JWT, 不会验证JWT签名
  verify_access_token: true
`

var TplConfYml = `log:
  # 日志级别 debug|info|warn|error|fatal
  level: "info"

resource:
%%TPL-CONF-YML-RESOURCE-DB%%%%TPL-CONF-YML-RESOURCE-ACCESS-TOKEN%%
%%TPL-CONF-YML-HTTP-SERVER%%
%%TPL-CONF-YML-GRPC-SERVER%%
monitor_server:
  # 监控服务监听地址, 如 ":8082"
  address: ""

constraint:
  # 时区 默认值: Asia/Shanghai
  timezone: "Asia/Shanghai"
`

package template

var TplManifestDTD = `<!-- project: 项目
name: 项目英文名
desc: 项目描述
go_module: 项目的go module名
go_version: 项目的go版本
server_type: 项目的服务类型 (HTTP|GRPC|ALL)
no_auth: 是否不需要鉴权 (true|false)
-->
<!ELEMENT project (usecases*, database*)>
<!ATTLIST project
    name CDATA #REQUIRED
    desc CDATA #REQUIRED
    go_module CDATA #REQUIRED
    go_version CDATA #REQUIRED
    server_type (HTTP|GRPC|ALL) #REQUIRED
    no_auth (true|false) "false"
>
<!-- usecases: 复杂用例 -->
<!ELEMENT usecases (usecase*)>
<!-- usecase: 用例
name: 用例名, 保留词: usecase, biz
desc: 用例描述
-->
<!ELEMENT usecase EMPTY>
<!ATTLIST usecase
    name CDATA #REQUIRED
    desc CDATA #REQUIRED
>
<!-- database: 数据库
name: 库名
charset: 字符集
collate: 排序规则
-->
<!ELEMENT database (table+)>
<!ATTLIST database
    name CDATA #REQUIRED
    type (MySQL|TiDB|PostgreSQL) #REQUIRED
    charset CDATA "utf8mb4"
    collate CDATA #IMPLIED
>
<!-- table: 数据库表
name: 表名, 保留词: base
type: 表类型
    META: 元数据表
    DATA: 实体表
    BR: 多对多的二元关系表
    RL: 一对多关系表
    RI: 倒排索引表
charset: 字符集
collate: 排序规则
comment: 表描述
handler: 是否生成该表的handler
-->
<!ELEMENT table (column+, index*, unique*, fulltext*)>
<!ATTLIST table
    name CDATA #REQUIRED
    type (META|DATA|BR|RL|RI) #REQUIRED
    charset CDATA #IMPLIED
    collate CDATA #IMPLIED
    comment CDATA #REQUIRED
    handler (true|false) "false"
>
<!-- column: 表字段
name: 字段名
type: 字段类型
    _ID: 默认的表主键类型 bigint(20, unsigned) primary auto_increment not null
    _FID: 默认的外链类型 bigint(20, unsigned)
    _SMT: META表的语义字段 CHAR(16)，可自定义长度 (deprecated)
    _SMT2: META表的语义字段 VARCHAR(16)，可自定义长度 (推荐)
    _DESC: META表的描述字段 VARCHAR(128)，可自定义长度
    _BOOL: MySQL/TiDB:布尔类型 TINYINT(1) | PostgreSQL: BOOL
    _ME_CREATE: 创建者
    _ME_UPDATE: 更新者
    _ME_DELETE: 删除者, 必须DEFAULT NULL
    _TIME_CREATE: 创建时间
    _TIME_UPDATE: 更新时间
    _TIME_DELETE: 删除时间(datetime), NULL代表未删除, 必须DEFAULT NULL (deprecated, 不适用于跟其他字段做联合唯一索引)
    _TIME_DELETE2: 删除时间(bigint,秒级时间戳), 0代表未删除, 适用于跟其他字段做联合唯一索引
    TIMETZ: time with timezone (only for PostgreSQL)
    TIMESTAMPTZ: timestamp with timezone (only for PostgreSQL)
    JSONB: jsonb (only for PostgreSQL)
    TINYBLOB|BLOB|MEDIUMBLOB|LONGBLOB: 变长二进制类型 (only for MySQL/TiDB)
    BINARY|VARBINARY: 定长二进制类型 (only for MySQL/TiDB)
    BYTEA: bytea (only for PostgreSQL)
length: 字段长度/数值最大位数
decimal: 仅对decimal类型的字段有用, 表示小数点右边的位数
unsigned: 是否无符号数值类型
required: 是否Not Null
primary_key: 是否主键，限一个
index: 是否对字段加索引
unique: 是否对字段加唯一索引
fulltext: 是否对字段加全文索引(ngram) (仅对CHAR/VARCHAR/TEXT类型的字段有用) (目前仅对MySQL有用)
comment: 字段描述
insert_default: 在插入记录时该字段的默认值
default_empty: 在插入记录时该字段的默认值为空字符串 (仅对CHAR/VARCHAR/TEXT类型的字段有用)
ri_key: 是否倒排索引的key
hidden: 在标准handler里是否对前端隐藏
filter: 是否支持filter
order: 是否支持order
list:  是否在列表页显示
alter: 是否允许修改
-->
<!ELEMENT column (foreign_key*)>
<!ATTLIST column
    name CDATA #REQUIRED
    type (CHAR|VARCHAR|TEXT|TINYINT|SMALLINT|MEDIUMINT|INT|BIGINT|FLOAT|DOUBLE|DECIMAL|DATE|DATETIME|TIMESTAMP|TIME|TIMETZ|TIMESTAMPTZ|JSON|JSONB|TINYBLOB|BLOB|MEDIUMBLOB|LONGBLOB|BINARY|VARBINARY|BYTEA|_ID|_FID|_BOOL|_SMT|_ME_CREATE|_ME_UPDATE|_ME_DELETE|_TIME_CREATE|_TIME_UPDATE|_TIME_DELETE|_SMT2|_DESC|_TIME_DELETE2) #REQUIRED
    length CDATA #IMPLIED
    decimal CDATA #IMPLIED
    unsigned (true|false) "false"
    required (true|false) "false"
    primary_key (true|false) "false"
    index (true|false) "false"
    unique (true|false) "false"
    fulltext (true|false) "false"
    comment CDATA #IMPLIED
    insert_default CDATA #IMPLIED
    default_empty (true|false) "false"
    ri_key (true|false) "false"
    hidden (true|false) "false"
    filter (true|false) "false"
    order (true|false) "false"
    list (true|false) "false"
    alter (true|false) "false"
>
<!-- index: 联合索引，至少要有两个字段以上-->
<!ELEMENT index (index_column, index_column+)>
<!-- unique: 联合唯一索引，至少要有两个字段以上-->
<!ELEMENT unique (index_column, index_column+)>
<!-- fulltext: 联合全文搜索索引，至少要有两个字段以上-->
<!ELEMENT fulltext (index_column, index_column+)>
<!-- index_column: 联合索引字段
name: 字段名
-->
<!ELEMENT index_column EMPTY>
<!ATTLIST index_column name CDATA #REQUIRED>
<!-- foreign_key: 外链
table: 外链的目标数据库表
counter: 计数对应的表字段名
ri_raw: 倒排索引key的原始字段名
auto_remove: 当链接的目标record被删除时, 是否自动删除当前record，仅对RL表有效
is_main: 对于RL表，标识此外键是否指向主表 (true|false)
-->
<!ELEMENT foreign_key EMPTY>
<!ATTLIST foreign_key
    table CDATA #REQUIRED
    counter CDATA #IMPLIED
    ri_raw CDATA #IMPLIED
    auto_remove (true|false) "false"
    is_main (true|false) "false"
>
`

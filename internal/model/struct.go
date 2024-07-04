package model

type ServerType string // 服务类型
var (
	ServerType_HTTP ServerType = "HTTP"
	ServerType_GRPC ServerType = "GRPC"
	ServerType_ALL  ServerType = "ALL"
)

type DatabaseType string // 数据库类型

var (
	DBType_MYSQL DatabaseType = "MySQL"
	DBType_TIDB  DatabaseType = "TiDB"
	DBType_PG    DatabaseType = "PostgreSQL"
)

type TableType string // 表类型

var (
	TableType_META TableType = "META" // 元数据表
	TableType_DATA TableType = "DATA" // 实体表
	TableType_BR   TableType = "BR"   // 二元关系表
	TableType_RL   TableType = "RL"   // 一对多关系表
	TableType_RI   TableType = "RI"   // 倒排索引表
)

var AvailableTableTypes = map[TableType]bool{
	TableType_META: true,
	TableType_DATA: true,
	TableType_BR:   true,
	TableType_RL:   true,
	TableType_RI:   true,
}

type ColumnType string    // 字段基础类型
type ColumnExtType string // 字段扩展类型

var (
	// 基础数据库字段类型
	ColumnType_CHAR      ColumnType = "CHAR"
	ColumnType_VARCHAR   ColumnType = "VARCHAR"
	ColumnType_TEXT      ColumnType = "TEXT"
	ColumnType_JSON      ColumnType = "JSON"
	ColumnType_JSONB     ColumnType = "JSONB"
	ColumnType_TINYINT   ColumnType = "TINYINT"
	ColumnType_SMALLINT  ColumnType = "SMALLINT"
	ColumnType_MEDIUMINT ColumnType = "MEDIUMINT"
	ColumnType_INT       ColumnType = "INT"
	ColumnType_BIGINT    ColumnType = "BIGINT"
	ColumnType_FLOAT     ColumnType = "FLOAT"
	ColumnType_DOUBLE    ColumnType = "DOUBLE"
	ColumnType_DECIMAL   ColumnType = "DECIMAL"
	ColumnType_DATE      ColumnType = "DATE"
	ColumnType_DATETIME  ColumnType = "DATETIME"
	ColumnType_TIMESTAMP ColumnType = "TIMESTAMP"
	ColumnType_TIME      ColumnType = "TIME"
	// for mysql/TiDB only
	ColumnType_TINYBLOB   ColumnType = "TINYBLOB"
	ColumnType_BLOB       ColumnType = "BLOB"
	ColumnType_MEDIUMBLOB ColumnType = "MEDIUMBLOB"
	ColumnType_LONGBLOB   ColumnType = "LONGBLOB"
	ColumnType_BINARY     ColumnType = "BINARY"
	ColumnType_VARBINARY  ColumnType = "VARBINARY"
	// for pg only
	ColumnType_BOOL        ColumnType = "BOOL"
	ColumnType_TIMETZ      ColumnType = "TIMETZ"
	ColumnType_TIMESTAMPTZ ColumnType = "TIMESTAMPTZ"
	ColumnType_BYTEA       ColumnType = "BYTEA"

	// 自定义字段扩展类型
	ColumnExtType_ID           ColumnExtType = "_ID"           // 默认的表主键类型 bigint(20, unsigned) primary auto_increment not null
	ColumnExtType_FID          ColumnExtType = "_FID"          // 默认的外链类型 bigint(20, unsigned)
	ColumnExtType_BOOL         ColumnExtType = "_BOOL"         // 布尔类型 tinyint(1)
	ColumnExtType_SMT          ColumnExtType = "_SMT"          // META表的语义字段 CHAR(16)，可自定义长度
	ColumnExtType_SMT2         ColumnExtType = "_SMT2"         // META表的语义字段 VARCHAR(16)，可自定义长度
	ColumnExtType_DESC         ColumnExtType = "_DESC"         // META表的描述字段 VARCHAR(128)，可自定义长度
	ColumnExtType_ME_CREATE    ColumnExtType = "_ME_CREATE"    // 创建者
	ColumnExtType_ME_UPDATE    ColumnExtType = "_ME_UPDATE"    // 更新者
	ColumnExtType_ME_DELETE    ColumnExtType = "_ME_DELETE"    // 删除者
	ColumnExtType_TIME_CREATE  ColumnExtType = "_TIME_CREATE"  // 创建时间
	ColumnExtType_TIME_UPDATE  ColumnExtType = "_TIME_UPDATE"  // 更新时间
	ColumnExtType_TIME_DELETE  ColumnExtType = "_TIME_DELETE"  // 删除时间，NULL代表未删除
	ColumnExtType_TIME_DELETE2 ColumnExtType = "_TIME_DELETE2" // 删除时间，0代表未删除
)

var AvailableColumnTypes = map[ColumnType]bool{
	ColumnType_CHAR:        true,
	ColumnType_VARCHAR:     true,
	ColumnType_TEXT:        true,
	ColumnType_JSON:        true,
	ColumnType_JSONB:       true,
	ColumnType_TINYINT:     true,
	ColumnType_SMALLINT:    true,
	ColumnType_MEDIUMINT:   true,
	ColumnType_INT:         true,
	ColumnType_BIGINT:      true,
	ColumnType_FLOAT:       true,
	ColumnType_DOUBLE:      true,
	ColumnType_DECIMAL:     true,
	ColumnType_DATE:        true,
	ColumnType_DATETIME:    true,
	ColumnType_TIMESTAMP:   true,
	ColumnType_TIME:        true,
	ColumnType_BOOL:        true,
	ColumnType_TIMETZ:      true,
	ColumnType_TIMESTAMPTZ: true,
	ColumnType_TINYBLOB:    true,
	ColumnType_BLOB:        true,
	ColumnType_MEDIUMBLOB:  true,
	ColumnType_LONGBLOB:    true,
	ColumnType_BINARY:      true,
	ColumnType_VARBINARY:   true,
	ColumnType_BYTEA:       true,
}

var ColumnExtTypeToColumnType = map[ColumnExtType]ColumnType{
	ColumnExtType_ID:           ColumnType_BIGINT,
	ColumnExtType_FID:          ColumnType_BIGINT,
	ColumnExtType_BOOL:         ColumnType_BOOL,
	ColumnExtType_SMT:          ColumnType_CHAR,
	ColumnExtType_SMT2:         ColumnType_VARCHAR,
	ColumnExtType_DESC:         ColumnType_VARCHAR,
	ColumnExtType_ME_CREATE:    ColumnType_BIGINT,
	ColumnExtType_ME_UPDATE:    ColumnType_BIGINT,
	ColumnExtType_ME_DELETE:    ColumnType_BIGINT,
	ColumnExtType_TIME_CREATE:  ColumnType_DATETIME,
	ColumnExtType_TIME_UPDATE:  ColumnType_DATETIME,
	ColumnExtType_TIME_DELETE:  ColumnType_DATETIME,
	ColumnExtType_TIME_DELETE2: ColumnType_BIGINT,
}

var IntegerColumnTypes = map[ColumnType]bool{
	ColumnType_TINYINT:   true,
	ColumnType_SMALLINT:  true,
	ColumnType_MEDIUMINT: true,
	ColumnType_INT:       true,
	ColumnType_BIGINT:    true,
}

var NumericColumnTypes = map[ColumnType]bool{
	ColumnType_TINYINT:   true,
	ColumnType_SMALLINT:  true,
	ColumnType_MEDIUMINT: true,
	ColumnType_INT:       true,
	ColumnType_BIGINT:    true,
	ColumnType_FLOAT:     true,
	ColumnType_DOUBLE:    true,
	ColumnType_DECIMAL:   true,
}

var StringColumnTypes = map[ColumnType]bool{
	ColumnType_CHAR:    true,
	ColumnType_VARCHAR: true,
	ColumnType_TEXT:    true,
	ColumnType_JSON:    true,
	ColumnType_JSONB:   true,
}

var TimeColumnTypes = map[ColumnType]bool{
	ColumnType_DATE:        true,
	ColumnType_DATETIME:    true,
	ColumnType_TIMESTAMP:   true,
	ColumnType_TIME:        true,
	ColumnType_TIMETZ:      true,
	ColumnType_TIMESTAMPTZ: true,
}

// support full text search
var FullTextColumnTypes = map[ColumnType]bool{
	ColumnType_CHAR:    true,
	ColumnType_VARCHAR: true,
	ColumnType_TEXT:    true,
}

// 二进制类型
var BinaryColumnTypes = map[ColumnType]bool{
	ColumnType_TINYBLOB:   true,
	ColumnType_BLOB:       true,
	ColumnType_MEDIUMBLOB: true,
	ColumnType_LONGBLOB:   true,
	ColumnType_BINARY:     true,
	ColumnType_VARBINARY:  true,
	ColumnType_BYTEA:      true,
}

// 不支持默认值
var NoDefaultColTypes = map[ColumnType]bool{
	ColumnType_TEXT:       true,
	ColumnType_TINYBLOB:   true,
	ColumnType_BLOB:       true,
	ColumnType_MEDIUMBLOB: true,
	ColumnType_LONGBLOB:   true,
}

var DefaultEmptyColTypes = map[ColumnType]bool{
	ColumnType_CHAR:    true,
	ColumnType_VARCHAR: true,
	ColumnType_TEXT:    true,
}

// 项目
type Project struct {
	Name       string     `xml:"name,attr"`        // 项目英文名
	Desc       string     `xml:"desc,attr"`        // 项目描述
	GoModule   string     `xml:"go_module,attr"`   // 项目的go module名
	GoVersion  string     `xml:"go_version,attr"`  // 项目的go版本
	ServerType ServerType `xml:"server_type,attr"` // 项目的server类型
	NoAuth     bool       `xml:"no_auth,attr"`     // 是否不需要鉴权
	Domain     *Domain    `xml:"usecases"`         // 复杂用例
	Database   *Database  `xml:"database"`         // 数据库
}

// 复杂用例
type Domain struct {
	Usecases []*Usecase `xml:"usecase"` // 非匿名用例

	Project              *Project            `xml:"-" json:"-"` // 所属项目
	UsecaseNameToUsecase map[string]*Usecase `xml:"-" json:"-"` // 用例名到用例的映射
}

// 用例
type Usecase struct {
	Name   string   `xml:"name,attr"` // 用例名
	Desc   string   `xml:"desc,attr"` // 用例描述
	Tables []*Table `xml:"table"`     // 关联的数据库表

	Domain *Domain `xml:"-" json:"-"` // 所属领域
}

// 数据库
type Database struct {
	Name    string       `xml:"name,attr"`    // 库名
	Charset string       `xml:"charset,attr"` // 字符集
	Collate string       `xml:"collate,attr"` // 排序规则
	Type    DatabaseType `xml:"type,attr"`    // 数据库类型
	Tables  []*Table     `xml:"table"`        // 表

	Project          *Project          `xml:"-" json:"-"` // 所属项目
	TableNameToTable map[string]*Table `xml:"-" json:"-"` // 表名到表的映射
	IsSchemaByDB     bool              `xml:"-" json:"-"` // 是否是读数据库得到的schema
}

// 数据库表
type Table struct {
	Name       string      `xml:"name,attr"`    // 表名
	Type       TableType   `xml:"type,attr"`    // 表类型
	Charset    string      `xml:"charset,attr"` // 字符集
	Collate    string      `xml:"collate,attr"` // 排序规则
	Comment    string      `xml:"comment,attr"` // 表描述
	Columns    []*Column   `xml:"column"`       // 字段
	Indexes    []*Index    `xml:"index"`        // 联合索引
	Unique     []*Unique   `xml:"unique"`       // 联合唯一索引
	FullText   []*FullText `xml:"fulltext"`     // 联合全文搜索索引
	HasHandler bool        `xml:"handler,attr"` // 是否生成handler

	Database           *Database            `xml:"-" json:"-"` // 所属数据库
	PrimaryColumn      *Column              `xml:"-" json:"-"` // 主键字段
	ColNamesToIndex    map[string]*Index    // 字段名（逗号分隔&字符升序）到联合索引的映射
	ColNamesToUnique   map[string]*Unique   // 字段名（逗号分隔&字符升序）到联合唯一索引的映射
	ColNamesToFullText map[string]*FullText // 字段名（逗号分隔&字符升序）到联合全文搜索索引的映射
	ColNameToColumn    map[string]*Column   // 字段名到字段的映射
}

// 字段
type Column struct {
	Name          string        `xml:"name,attr"`           // 字段名
	Type          ColumnType    `xml:"type,attr"`           // 字段类型
	Length        int           `xml:"length,attr"`         // 字段长度/数值最大位数
	Decimal       int           `xml:"decimal,attr"`        // 仅对decimal类型的字段有用，表示小数点右边的位数
	IsUnsigned    bool          `xml:"unsigned,attr"`       // 是否无符号数值类型
	IsRequired    bool          `xml:"required,attr"`       // 是否Not Null
	IsPrimaryKey  bool          `xml:"primary_key,attr"`    // 是否主键，限一个
	IsIndex       bool          `xml:"index,attr"`          // 是否对字段加索引
	IsUnique      bool          `xml:"unique,attr"`         // 是否对字段加唯一索引
	IsFullText    bool          `xml:"fulltext,attr"`       // 是否对字段加全文搜索索引
	IsHidden      bool          `xml:"hidden,attr"`         // 在标准handler里是否对前端隐藏
	IsFilter      bool          `xml:"filter,attr"`         // 是否支持筛选
	IsOrder       bool          `xml:"order,attr"`          // 是否支持排序
	IsList        bool          `xml:"list,attr"`           // 是否在列表页显示
	IsAlterable   bool          `xml:"alter,attr"`          // 是否可修改
	Comment       string        `xml:"comment,attr"`        // 字段描述
	InsertDefault string        `xml:"insert_default,attr"` // 在插入记录时该字段的默认值
	DefaultEmpty  bool          `xml:"default_empty,attr"`  // 是否默认为false
	IsRiKey       bool          `xml:"ri_key,attr"`         // 是否倒排索引的key
	ForeignKeys   []*ForeignKey `xml:"foreign_key"`         // 外键，只有一个

	ExtType    ColumnExtType `xml:"-"`          // 字段扩展类型
	IndexAlias string        `xml:"-"`          // (唯一)索引别名
	Table      *Table        `xml:"-" json:"-"` // 所属表
}

// 外链
type ForeignKey struct {
	Table      string `xml:"table,attr"`       // 外链的目标数据库表
	Counter    string `xml:"counter,attr"`     // 计数对应的表字段名
	RiRaw      string `xml:"ri_raw,attr"`      // 倒排索引key的原始字段名
	AutoRemove bool   `xml:"auto_remove,attr"` // 当链接的目标record被删除时，是否自动删除当前record

	Primary string  `xml:"-"`          // 目标数据库表的主键字段名
	Alias   string  `xml:"-"`          // 外链的唯一别名
	Column  *Column `xml:"-" json:"-"` // 所属字段
}

// 联合索引
type Index struct {
	IndexColumns []*IndexColumn `xml:"index_column"`

	Alias string `xml:"-"`          // 联合索引别名
	Table *Table `xml:"-" json:"-"` // 所属表
}

// 联合唯一索引
type Unique struct {
	IndexColumns []*IndexColumn `xml:"index_column"`

	Alias string `xml:"-"`          // 联合索引别名
	Table *Table `xml:"-" json:"-"` // 所属表
}

// 联合全文搜索索引
type FullText struct {
	IndexColumns []*IndexColumn `xml:"index_column"`

	Alias string `xml:"-"`          // 联合索引别名
	Table *Table `xml:"-" json:"-"` // 所属表
}

// 联合索引的组成字段
type IndexColumn struct {
	Name string `xml:"name,attr"` // 字段名
}

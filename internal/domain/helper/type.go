package helper

import (
	"fmt"
	"strings"

	"meta-egg/internal/model"
)

func IsGoTypeNullable(goType string) bool {
	return strings.Contains(goType, "[]")
}

// 数据库字段类型跟go类型的映射
func GetGoType(col *model.Column) (string, error) {
	columnType := col.Type
	isUnsigned := col.IsUnsigned
	switch columnType {
	case model.ColumnType_BOOL:
		return "bool", nil
	case model.ColumnType_TINYINT:
		if col.ExtType == model.ColumnExtType_BOOL {
			return "bool", nil
		}
		if isUnsigned {
			return "uint8", nil
		}
		return "int8", nil
	case model.ColumnType_SMALLINT:
		if isUnsigned {
			return "uint16", nil
		}
		return "int16", nil
	case model.ColumnType_MEDIUMINT:
		if isUnsigned {
			return "uint32", nil
		}
		return "int32", nil
	case model.ColumnType_INT:
		if isUnsigned {
			return "uint32", nil
		}
		return "int32", nil
	case model.ColumnType_BIGINT:
		if isUnsigned {
			return "uint64", nil
		}
		return "int64", nil
	case model.ColumnType_FLOAT:
		return "float32", nil
	case model.ColumnType_DOUBLE:
		return "float64", nil
	case model.ColumnType_DECIMAL:
		return "float64", nil
	case model.ColumnType_CHAR,
		model.ColumnType_VARCHAR,
		model.ColumnType_TEXT:
		return "string", nil
	case model.ColumnType_JSON,
		model.ColumnType_JSONB,
		model.ColumnType_TINYBLOB,
		model.ColumnType_BLOB,
		model.ColumnType_MEDIUMBLOB,
		model.ColumnType_LONGBLOB,
		model.ColumnType_BINARY,
		model.ColumnType_VARBINARY,
		model.ColumnType_BYTEA:
		return "[]byte", nil
	case model.ColumnType_DATE,
		model.ColumnType_TIME,
		model.ColumnType_DATETIME,
		model.ColumnType_TIMESTAMP,
		model.ColumnType_TIMETZ,
		model.ColumnType_TIMESTAMPTZ:
		return "time.Time", nil
	default:
		return "unsupported", fmt.Errorf("unsupported column type: %v", columnType)
	}
}

// 数据库字段类型跟proto3里的字段类型的映射
func GetProto3ValueType(col *model.Column) (string, error) {
	columnType := col.Type
	isUnsigned := col.IsUnsigned
	switch columnType {
	case model.ColumnType_BOOL:
		return "bool", nil
	case model.ColumnType_TINYINT:
		if col.ExtType == model.ColumnExtType_BOOL {
			return "bool", nil
		}
		if isUnsigned {
			return "uint32", nil
		}
		return "int32", nil
	case model.ColumnType_SMALLINT,
		model.ColumnType_MEDIUMINT,
		model.ColumnType_INT:
		if isUnsigned {
			return "uint32", nil
		}
		return "int32", nil
	case model.ColumnType_BIGINT:
		if isUnsigned {
			return "uint64", nil
		}
		return "int64", nil
	case model.ColumnType_FLOAT:
		return "float", nil
	case model.ColumnType_DOUBLE:
		return "double", nil
	case model.ColumnType_DECIMAL:
		return "double", nil
	case model.ColumnType_CHAR,
		model.ColumnType_VARCHAR,
		model.ColumnType_TEXT:
		return "string", nil
	case model.ColumnType_JSON,
		model.ColumnType_JSONB,
		model.ColumnType_TINYBLOB,
		model.ColumnType_BLOB,
		model.ColumnType_MEDIUMBLOB,
		model.ColumnType_LONGBLOB,
		model.ColumnType_BINARY,
		model.ColumnType_VARBINARY,
		model.ColumnType_BYTEA:
		return "bytes", nil
	case model.ColumnType_DATETIME,
		model.ColumnType_TIMESTAMP,
		model.ColumnType_TIME,
		model.ColumnType_DATE,
		model.ColumnType_TIMETZ,
		model.ColumnType_TIMESTAMPTZ:
		return "string", nil
	default:
		return "unsupported", fmt.Errorf("unsupported column type: %v", columnType)
	}
}

func Proto3ValueType2GoType(proto3ValueType string) (string, error) {
	switch proto3ValueType {
	case "bool":
		return "bool", nil
	case "uint32":
		return "uint32", nil
	case "int32":
		return "int32", nil
	case "uint64":
		return "uint64", nil
	case "int64":
		return "int64", nil
	case "float":
		return "float32", nil
	case "double":
		return "float64", nil
	case "string":
		return "string", nil
	case "bytes":
		return "[]byte", nil
	default:
		return "unsupported", fmt.Errorf("unsupported proto3 value type: %v", proto3ValueType)
	}
}

// 数据库字段类型跟go类型的映射
func GetGoTypeForHandler(col *model.Column) string {
	columnType := col.Type
	isUnsigned := col.IsUnsigned
	var gotype string
	switch columnType {
	case model.ColumnType_BOOL:
		gotype = "bool"
	case model.ColumnType_TINYINT:
		if col.ExtType == model.ColumnExtType_BOOL {
			gotype = "bool"
		} else if isUnsigned {
			gotype = "uint8"
		} else {
			gotype = "int8"
		}
	case model.ColumnType_SMALLINT:
		if isUnsigned {
			gotype = "uint16"
		} else {
			gotype = "int16"
		}
	case model.ColumnType_MEDIUMINT:
		if isUnsigned {
			gotype = "uint32"
		} else {
			gotype = "int32"
		}
	case model.ColumnType_INT:
		if isUnsigned {
			gotype = "uint32"
		} else {
			gotype = "int32"
		}
	case model.ColumnType_BIGINT:
		if isUnsigned {
			gotype = "uint64"
		} else {
			gotype = "int64"
		}
	case model.ColumnType_FLOAT:
		gotype = "float32"
	case model.ColumnType_DOUBLE:
		gotype = "float64"
	case model.ColumnType_DECIMAL:
		gotype = "float64"
	case model.ColumnType_CHAR:
		gotype = "string"
	case model.ColumnType_VARCHAR:
		gotype = "string"
	case model.ColumnType_JSON,
		model.ColumnType_JSONB,
		model.ColumnType_TINYBLOB,
		model.ColumnType_BLOB,
		model.ColumnType_MEDIUMBLOB,
		model.ColumnType_LONGBLOB,
		model.ColumnType_BINARY,
		model.ColumnType_VARBINARY,
		model.ColumnType_BYTEA:
		gotype = "[]byte"
	case model.ColumnType_TEXT:
		gotype = "string"
	case model.ColumnType_DATE:
		gotype = "string" // 2006-01-02
	case model.ColumnType_TIME:
		gotype = "string" // 15:04:05
	case model.ColumnType_DATETIME:
		gotype = "string" // 2006-01-02 15:04:05
	case model.ColumnType_TIMESTAMP:
		gotype = "string" // 2006-01-02 15:04:05
	case model.ColumnType_TIMETZ:
		gotype = "string" // 15:04:05
	case model.ColumnType_TIMESTAMPTZ:
		gotype = "string" // 2006-01-02 15:04:05
	}

	if !col.IsRequired && !IsGoTypeNullable(gotype) {
		gotype = "*" + gotype
	}
	return gotype
}

func GetCommentForHandler(col *model.Column) string {
	comment := col.Comment
	columnType := col.Type
	switch columnType {
	case model.ColumnType_DATE:
		comment += ` 格式: 2006-01-02`
	case model.ColumnType_TIME:
		comment += ` 格式: 15:04:05`
	case model.ColumnType_DATETIME:
		comment += ` 格式: 2006-01-02 15:04:05`
	case model.ColumnType_TIMESTAMP:
		comment += ` 格式: 2006-01-02 15:04:05`
	case model.ColumnType_TIMETZ:
		comment += ` 格式: 15:04:05`
	case model.ColumnType_TIMESTAMPTZ:
		comment += ` 格式: 2006-01-02 15:04:05`
	}
	return comment
}

package helper

import (
	"fmt"
	"meta-egg/internal/model"
	"strings"
)

// like:
// binding:"omitempty,min=1,max=64"
func GetBinding(col *model.Column) string {
	var buf strings.Builder
	tags := make([]string, 0)
	if !col.IsRequired ||
		(model.DefaultEmptyColTypes[col.Type] && col.DefaultEmpty) ||
		(model.NumericColumnTypes[col.Type] && col.InsertDefault == "0") {
		tags = append(tags, "omitempty")
	} else if col.ExtType != model.ColumnExtType_BOOL && // bool default value is false
		col.Type != model.ColumnType_BOOL {
		tags = append(tags, "required")
	}
	switch col.Type {
	case model.ColumnType_TINYINT:
	case model.ColumnType_SMALLINT:
	case model.ColumnType_MEDIUMINT:
	case model.ColumnType_INT:
	case model.ColumnType_BIGINT:
	case model.ColumnType_FLOAT:
	case model.ColumnType_DOUBLE:
	case model.ColumnType_DECIMAL:
	case model.ColumnType_CHAR:
		tags = append(tags, fmt.Sprintf("max=%d", col.Length))
	case model.ColumnType_VARCHAR:
		tags = append(tags, fmt.Sprintf("max=%d", col.Length))
	case model.ColumnType_TEXT:
	case model.ColumnType_JSON:
	case model.ColumnType_JSONB:
	case model.ColumnType_TINYBLOB:
	case model.ColumnType_BLOB:
	case model.ColumnType_MEDIUMBLOB:
	case model.ColumnType_LONGBLOB:
	case model.ColumnType_BINARY:
	case model.ColumnType_VARBINARY:
	case model.ColumnType_BYTEA:
	case model.ColumnType_DATE:
		tags = append(tags, "datetime=2006-01-02")
	case model.ColumnType_TIME:
		tags = append(tags, "datetime=15:04:05")
	case model.ColumnType_DATETIME:
		tags = append(tags, "datetime=2006-01-02 15:04:05")
	case model.ColumnType_TIMESTAMP:
		tags = append(tags, "datetime=2006-01-02 15:04:05")
	case model.ColumnType_TIMETZ:
		tags = append(tags, "datetime=15:04:05")
	case model.ColumnType_TIMESTAMPTZ:
		tags = append(tags, "datetime=2006-01-02 15:04:05")
	}
	switch col.ExtType {
	case model.ColumnExtType_FID:
		tags = append(tags, "gte=1")
	case model.ColumnExtType_BOOL:
	}
	if len(tags) > 0 {
		buf.WriteString(" binding:\"")
		buf.WriteString(strings.Join(tags, ","))
		buf.WriteString("\"")
	}
	return buf.String()
}

// like:
// [(validate.rules).string = {min_len: 1, max_len: 8}]
func GetProto3ValidateRule(col *model.Column) string {
	var buf strings.Builder
	switch col.Type {
	case model.ColumnType_TINYINT:
	case model.ColumnType_SMALLINT:
	case model.ColumnType_MEDIUMINT:
	case model.ColumnType_INT:
	case model.ColumnType_BIGINT:
	case model.ColumnType_FLOAT:
	case model.ColumnType_DOUBLE:
	case model.ColumnType_DECIMAL:
	case model.ColumnType_CHAR:
		buf.WriteString(fmt.Sprintf("string = {max_len: %d}]", col.Length))
	case model.ColumnType_VARCHAR:
		buf.WriteString(fmt.Sprintf("string = {max_len: %d}]", col.Length))
	case model.ColumnType_TEXT:
	case model.ColumnType_JSON:
	case model.ColumnType_JSONB:
	case model.ColumnType_DATE:
		buf.WriteString(`string = {
        pattern: "^\\d{4}-\\d{2}-\\d{2}$",
    }]`)
	case model.ColumnType_TIME:
		buf.WriteString(`string = {
        pattern: "^\\d{2}:\\d{2}:\\d{2}$",
    }]`)
	case model.ColumnType_DATETIME:
		buf.WriteString(`string = {
        pattern: "^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}$",
    }]`)
	case model.ColumnType_TIMESTAMP:
		buf.WriteString(`string = {
        pattern: "^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}$",
    }]`)
	case model.ColumnType_TIMETZ:
		buf.WriteString(`string = {
        pattern: "^\\d{2}:\\d{2}:\\d{2}$",
    }]`)
	case model.ColumnType_TIMESTAMPTZ:
		buf.WriteString(`string = {
        pattern: "^\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2}$",
    }]`)
	}
	switch col.ExtType {
	case model.ColumnExtType_FID:
		buf.WriteString("uint64 = {gte: 1}]")
	case model.ColumnExtType_BOOL:
	}
	str := buf.String()
	if len(str) > 0 {
		return " [(validate.rules)." + str
	} else {
		return ""
	}
}

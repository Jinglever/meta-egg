package pgoperator

import (
	"fmt"
	"meta-egg/internal/model"

	jgpg "github.com/Jinglever/go-postgres"
)

func ConvertPgType2ModelColType(typ string) model.ColumnType {
	switch typ {
	case jgpg.CT_SMALLINT:
		return model.ColumnType_SMALLINT
	case jgpg.CT_INT:
		return model.ColumnType_INT
	case jgpg.CT_BIGINT:
		return model.ColumnType_BIGINT
	case jgpg.CT_DECIMAL:
		return model.ColumnType_DECIMAL
	case jgpg.CT_FLOAT:
		return model.ColumnType_FLOAT
	case jgpg.CT_DOUBLE:
		return model.ColumnType_DOUBLE
	case jgpg.CT_TEXT:
		return model.ColumnType_TEXT
	case jgpg.CT_CHAR:
		return model.ColumnType_CHAR
	case jgpg.CT_VARCHAR:
		return model.ColumnType_VARCHAR
	case jgpg.CT_JSON:
		return model.ColumnType_JSON
	case jgpg.CT_JSONB:
		return model.ColumnType_JSONB
	case jgpg.CT_DATE:
		return model.ColumnType_DATE
	case jgpg.CT_TIME:
		return model.ColumnType_TIME
	case jgpg.CT_TIMESTAMP:
		return model.ColumnType_DATETIME
	case jgpg.CT_BOOL:
		return model.ColumnType_BOOL
	case jgpg.CT_TIMETZ:
		return model.ColumnType_TIMETZ
	case jgpg.CT_TIMESTAMPTZ:
		return model.ColumnType_TIMESTAMPTZ
	case jgpg.CT_BYTEA:
		return model.ColumnType_BYTEA
	}
	return model.ColumnType(typ)
}

func GetPgType(col *model.Column) string {
	typ := string(col.Type)
	switch col.Type {
	case model.ColumnType_SMALLINT:
		typ = jgpg.CT_SMALLINT
	case model.ColumnType_INT:
		typ = jgpg.CT_INT
	case model.ColumnType_BIGINT:
		typ = jgpg.CT_BIGINT
	case model.ColumnType_DECIMAL:
		typ = jgpg.CT_DECIMAL
	case model.ColumnType_FLOAT:
		typ = jgpg.CT_FLOAT
	case model.ColumnType_DOUBLE:
		typ = jgpg.CT_DOUBLE
	case model.ColumnType_TEXT:
		typ = jgpg.CT_TEXT
	case model.ColumnType_CHAR:
		typ = jgpg.CT_CHAR
	case model.ColumnType_VARCHAR:
		typ = jgpg.CT_VARCHAR
	case model.ColumnType_JSON:
		typ = jgpg.CT_JSON
	case model.ColumnType_JSONB:
		typ = jgpg.CT_JSONB
	case model.ColumnType_DATE:
		typ = jgpg.CT_DATE
	case model.ColumnType_TIME:
		typ = jgpg.CT_TIME
	case model.ColumnType_DATETIME:
		typ = jgpg.CT_TIMESTAMP
	case model.ColumnType_TIMESTAMP:
		typ = jgpg.CT_TIMESTAMP
	case model.ColumnType_TIMETZ:
		typ = jgpg.CT_TIMETZ
	case model.ColumnType_TIMESTAMPTZ:
		typ = jgpg.CT_TIMESTAMPTZ
	case model.ColumnType_TINYINT:
		typ = jgpg.CT_SMALLINT
	case model.ColumnType_BOOL:
		typ = jgpg.CT_BOOL
	case model.ColumnType_BYTEA:
		typ = jgpg.CT_BYTEA
	}
	if typ == jgpg.CT_DECIMAL {
		typ += fmt.Sprintf("(%d,%d)", col.Length, col.Decimal)
	} else if col.Length > 0 &&
		(model.StringColumnTypes[col.Type] ||
			model.BinaryColumnTypes[col.Type]) {
		typ += fmt.Sprintf("(%d)", col.Length)
	}
	return typ
}

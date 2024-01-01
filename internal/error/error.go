package cerror

import "fmt"

var (
	ErrDatabaseNotExists       = NewError(1001, "database not exists")
	ErrUnsupportedColumnType   = NewError(1002, "unsupported column type")
	ErrDBNotConnected          = NewError(1003, "database not connected")
	ErrUnsupportedDatabaseType = NewError(1004, "unsupported database type")
)

func NewError(code int, msg string) error {
	return fmt.Errorf("code: %d, msg: %s", code, msg)
}

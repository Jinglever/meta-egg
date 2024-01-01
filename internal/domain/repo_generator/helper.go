package repogen

import (
	"fmt"
	"strings"

	"meta-egg/internal/domain/helper"
	"meta-egg/internal/model"
)

// 填充当前时间
func setModelTimeNowByCol(body *strings.Builder, col *model.Column, imports *map[string]bool) {
	if col != nil {
		if !col.IsRequired {
			body.WriteString(fmt.Sprintf(`
			%s := time.Now()
			m.%s = &%s`, col.Name, helper.GetTableColName(col.Name), col.Name))
		} else {
			body.WriteString(fmt.Sprintf(`
			m.%s = time.Now()`, helper.GetTableColName(col.Name)))
		}
		if !(*imports)["time"] {
			(*imports)["time"] = true
		}
	}
}

// 填充指定值
func setModelValueByCol(body *strings.Builder, col *model.Column, val string) {
	if col != nil {
		if !col.IsRequired {
			body.WriteString(fmt.Sprintf(`
			%s := %s
			m.%s = &%s`, col.Name, val, helper.GetTableColName(col.Name), col.Name))
		} else {
			body.WriteString(fmt.Sprintf(`
			m.%s = %s`, helper.GetTableColName(col.Name), val))
		}
	}
}

// 填充当前时间
func setMapTimeNowByCol(body *strings.Builder, col *model.Column, imports *map[string]bool) {
	if col != nil {
		body.WriteString(fmt.Sprintf(`
			mp["%s"] = time.Now()`, col.Name))
		if !(*imports)["time"] {
			(*imports)["time"] = true
		}
	}
}

// 填充指定值
func setMapValueByCol(body *strings.Builder, col *model.Column, val string) {
	if col != nil {
		body.WriteString(fmt.Sprintf(`
			mp["%s"] = %s`, col.Name, val))
	}
}

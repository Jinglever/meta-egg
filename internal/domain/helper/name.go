package helper

import (
	"strings"

	jgstr "github.com/Jinglever/go-string"
	"github.com/gertd/go-pluralize"
)

// 表名/字段名到go结构体名的映射
func GetTableColName(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	parts := strings.Split(name, "_")
	var buf strings.Builder
	for _, part := range parts {
		if part == "id" {
			part = "ID"
		}
		buf.WriteString(jgstr.Ucfirst(part))
	}
	return buf.String()
}

func GetEnvPrefix(projName string) string {
	projName = strings.ReplaceAll(projName, "-", "_")
	if strings.Contains(projName, "_") {
		// split by '_' then take the first char of each part
		parts := strings.Split(projName, "_")
		var buf strings.Builder
		for _, part := range parts {
			buf.WriteString(strings.ToUpper(part[0:1]))
		}
		projName = buf.String()
	}
	return strings.ToUpper(projName)
}

// 全小写，下划线分隔
func GetDirName(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

// 全小写，无分隔
func GetPkgName(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), "_", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "", "")
	return name
}

// 首字母大写，驼峰命名
func GetStructName(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	parts := strings.Split(name, "_")
	var buf strings.Builder
	for _, part := range parts {
		buf.WriteString(jgstr.Ucfirst(part))
	}
	return buf.String()
}

// 全小写，中横线分隔
func NormalizeProjectName(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

// 首字母小写，驼峰命名
func GetVarName(name string) string {
	return jgstr.Lcfirst(GetStructName(name))
}

// 全小写，下划线分隔，改成复数形式
func GetURIName(name string) string {
	name = strings.ReplaceAll(strings.ToLower(name), "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	// 取复数形式
	plz := pluralize.NewClient()
	return plz.Plural(name)
}

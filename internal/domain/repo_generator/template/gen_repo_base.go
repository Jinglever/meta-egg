package template

import "meta-egg/internal/domain/helper"

var TplGenRepoBase string = helper.PH_META_EGG_HEADER + `
package repo

import (
)

const (
	CreateBatchNum = 100 // 批量插入的数量
)
`

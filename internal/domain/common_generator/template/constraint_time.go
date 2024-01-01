package template

import "meta-egg/internal/domain/helper"

var TplConstraintTime = helper.PH_META_EGG_HEADER + `
package constraint

const (
	// 日期格式
	DateFormat = "2006-01-02"
	// 分钟级时间格式
	MinuteTimeFormat = "2006-01-02 15:04"
	// 秒级时间格式
	SecondTimeFormat = "2006-01-02 15:04:05"
	// 时分格式
	HourMinuteFormat = "15:04"
	// 时分秒格式
	HourMinuteSecondFormat = "15:04:05"
)
`

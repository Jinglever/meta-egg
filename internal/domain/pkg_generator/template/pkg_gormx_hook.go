package template

import "meta-egg/internal/domain/helper"

var TplPkgGormxHook string = helper.PH_META_EGG_HEADER + `
package gormx

import "time"

func CorrectTimezone(t time.Time) time.Time {
	return time.Date(t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(), time.Local)
}
`

package template

import "meta-egg/internal/domain/helper"

var TplIntegrateHelper = helper.PH_META_EGG_HEADER + `
package integrate

import (
	"%%GO-MODULE%%/internal/common/resource"
	"%%GO-MODULE%%/internal/config"
)

func GetResource() *resource.Resource {
	// load config
	if err := config.LoadConfig("../../configs/conf-local.yml"); err != nil {
		panic(err)
	}
	rsrc, err := resource.InitResource(config.GetResourceConfig())
	if err != nil {
		panic(err)
	}
	return rsrc
}
`

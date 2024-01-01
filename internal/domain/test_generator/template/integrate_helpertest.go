package template

import "meta-egg/internal/domain/helper"

var TplIntegrateHelperTest = helper.PH_META_EGG_HEADER + `
package integrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetResource(t *testing.T) {
	rsrc := GetResource()
	assert.NotNil(t, rsrc)
}
`

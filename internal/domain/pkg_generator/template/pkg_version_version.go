package template

import "meta-egg/internal/domain/helper"

var TplPkgVersionVersion string = helper.PH_META_EGG_HEADER + `
package version

import (
	"flag"
	"fmt"
)

var (
	BuildTS   = "None"
	GitHash   = "None"
	GitBranch = "None"
	GitTag    = "None"
	GitDirty  = "None"
	Debug     = "None"
)

var (
	PrintVersion = flag.Bool("version", false, "print the version of this build")
)

// GetVersion prints build version.
func GetVersion() string {
	version := ""
	if GitTag != "" {
		version += GitTag
	} else if GitBranch != "" {
		version += GitBranch
	}
	if GitHash != "" {
		h := GitHash
		if len(h) > 7 { //nolint
			h = h[:7]
		}
		version += "-" + h
	}
	if Debug == "true" {
		version += "-debug"
	}
	if GitDirty == "true" {
		version += "-dirty"
	}
	return version
}

func Printer() {
	fmt.Println("Version:          ", GetVersion())
	fmt.Println("Git Commit:       ", GitHash)
	fmt.Println("Build Time (UTC): ", BuildTS)
}
`

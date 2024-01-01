package template

import "meta-egg/internal/domain/helper"

var TplInternalServerMonitorRouter string = helper.PH_META_EGG_HEADER + `
package monitor

import (
	"net/http"
	"net/http/pprof"
)

func (s *Server) initRouter() {
	router := http.NewServeMux()
	router.HandleFunc("/debug/pprof", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
	router.HandleFunc("/debug/pprof/{action}", pprof.Index)

	s.Router = router
}
`

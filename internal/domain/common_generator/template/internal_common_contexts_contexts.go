package template

import "meta-egg/internal/domain/helper"

var TplInternalCommonContextsContexts string = helper.PH_META_EGG_HEADER + `
package contexts

import (
	"context"
	"encoding/json"

	log "%%GO-MODULE%%/pkg/log"
)

type KeyME struct{}
type KeyLogger struct{}

type ME struct {
	ID uint64 ` + "`" + `json:"id"` + "`" + ` // ID of the user
}

func (me *ME) MustToJSON() string {
	s, err := json.Marshal(me)
	if err != nil {
		log.Fatalf("ME.ToJSON Err: %v", err)
	}
	return string(s)
}

func SetME(ctx context.Context, me ME) context.Context {
	return context.WithValue(ctx, KeyME{}, me)
}

func GetME(ctx context.Context) (ME, bool) {
	me, ok := ctx.Value(KeyME{}).(ME)
	return me, ok
}

func DelME(ctx context.Context) context.Context {
	return context.WithValue(ctx, KeyME{}, nil)
}

func SetLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, KeyLogger{}, logger)
}

// always return a logger
func GetLogger(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value(KeyLogger{}).(*log.Logger)
	if !ok {
		return log.WithFields(log.Fields{})
	}
	return logger
}

func DelLogger(ctx context.Context) context.Context {
	return context.WithValue(ctx, KeyLogger{}, nil)
}
`

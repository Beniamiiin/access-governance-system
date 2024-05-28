package di

import (
	prettyconsole "github.com/thessem/zap-prettyconsole"
	"go.uber.org/zap"
)

func NewLogger() *zap.SugaredLogger {
	return prettyconsole.NewLogger(zap.DebugLevel).Sugar()
}

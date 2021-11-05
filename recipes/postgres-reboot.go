package recipes

import (
	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/cmdctx"
)

type PostgresReboot struct {
	cmdCtx *cmdctx.CmdContext
	App    *api.App
}

func NewPostgresReboot(cmdCtx *cmdctx.CmdContext, app *api.App) *PostgresReboot {
	return &PostgresReboot{cmdCtx: cmdCtx, App: app}
}

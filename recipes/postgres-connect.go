package recipes

import (
	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/cmdctx"
	"github.com/superfly/flyctl/internal/recipe"
)

func PostgresConnectRecipe(cmdctx *cmdctx.CmdContext, app *api.App) error {
	ctx := cmdctx.Command.Context()

	recipe, err := recipe.NewRecipe(ctx, app)
	if err != nil {
		return err
	}

	machines, err := recipe.Client.API().ListMachines(ctx, app.ID, "started")
	if err != nil {
		return err
	}

	_, err = recipe.RunSSHAttachOperation(ctx, machines[0], PG_CONNECT)
	if err != nil {
		return err
	}

	return nil
}

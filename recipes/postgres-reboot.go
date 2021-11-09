package recipes

import (
	"context"

	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/internal/recipes"
)

func PostgresRebootRecipe(ctx context.Context, app *api.App) error {

	recipe, err := recipes.NewRecipe(ctx, app)
	if err != nil {
		return err
	}

	instances, err := recipe.Agent.Instances(ctx, &app.Organization, app.Name)
	stateOperations, err := recipe.RunOperation(instances.Addresses, ".flyd/bin/role")
	if err != nil {
		return err
	}

	roleMap := map[string][]string{}
	for _, stateOp := range stateOperations {
		roleMap[stateOp.Result] = append(roleMap[stateOp.Result], stateOp.Addr)
	}

	_, err = recipe.RunOperation(roleMap["replica"], ".flyd/bin/restart")
	if err != nil {
		return err
	}

	_, err = recipe.RunOperation(roleMap["leader"], ".flyd/bin/trigger-failover")
	if err != nil {
		return err
	}

	_, err = recipe.RunOperation(roleMap["leader"], ".flyd/bin/restart")
	if err != nil {
		return err
	}

	return nil
}

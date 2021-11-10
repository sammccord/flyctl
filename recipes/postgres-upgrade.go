package recipes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/internal/recipes"
)

type PostgresUpgradeConfig struct {
	AppName        string
	TargetImageRef string
}

func PostgresUpgradeRecipe(ctx context.Context, app *api.App, image string) error {

	recipe, err := recipes.NewRecipe(ctx, app)
	if err != nil {
		return err
	}

	// Fetch machines
	machines, err := recipe.Client.API().ListMachines(ctx, app.Name, "")
	if err != nil {
		return err
	}

	var roleMap map[string][]*api.Machine

	// Collect PG role information from each machine
	for _, machine := range machines {
		stateOp, err := recipe.RunSSHOperation(ctx, machine, PG_ROLE_SCRIPT)
		if err != nil {
			return err
		}
		roleMap[stateOp.Result] = append(roleMap[stateOp.Result], stateOp.Machine)
	}

	// Stop Replica/Launch new machine with  new image to replace it.
	for _, machine := range roleMap["replica"] {
		if err = replaceMachine(ctx, recipe, app, machine); err != nil {
			return err
		}
	}

	// Trigger failover and replace old leader.
	for _, machine := range roleMap["leader"] {
		_, err = recipe.RunSSHOperation(ctx, machine, PG_FAILOVER_SCRIPT)
		if err != nil {
			return err
		}

		if err = replaceMachine(ctx, recipe, app, machine); err != nil {
			return err
		}
	}

	return nil
}

func replaceMachine(ctx context.Context, recipe *recipes.Recipe, app *api.App, machine *api.Machine) error {

	stopEndpoint := fmt.Sprintf("/v1/machines/%s/stop", machine.ID)
	_, err := recipe.RunHTTPOperation(ctx, machine, http.MethodPost, stopEndpoint)
	if err != nil {
		return err
	}

	destroyEndpoint := fmt.Sprintf("/v1/machines/%s/", machine.ID)
	_, err = recipe.RunHTTPOperation(ctx, machine, http.MethodDelete, destroyEndpoint)
	if err != nil {
		return err
	}

	newConfig := machine.Config
	newConfig["image"] = "flyio/postgres:14"

	launchInput := api.LaunchMachineInput{
		AppID:  app.ID,
		Region: machine.Region,
		Config: &newConfig,
	}

	fmt.Printf("Launching new machine to replace %s\n", machine.ID)

	m, _, err := recipe.Client.API().LaunchMachine(ctx, launchInput)
	if err != nil {
		return err
	}

	WaitForMachineState(ctx, recipe.Client, app.ID, m.ID, "started")

	return nil
}

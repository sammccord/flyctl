package recipes

import (
	"context"
	"fmt"

	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/internal/recipes"
)

type PostgresUpgradeConfig struct {
	AppName        string
	TargetImageRef string
}

type PGMachine struct {
	Machine *api.Machine
	Role    string
	IP      string
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

	var pgMachines []PGMachine

	// Resolve PG role per machine.
	for _, machine := range machines {
		ip := machine.IPs.Nodes[0].IP
		stateOperation, err := recipe.RunOperation([]string{ip}, ".flyd/bin/role")
		if err != nil {
			return err
		}

		pgMachines = append(pgMachines, PGMachine{
			Machine: machine,
			IP:      ip,
			Role:    stateOperation[0].Result,
		})
	}

	// Stop Replica/Launch new machine with  new image to replace it.
	for _, pgM := range pgMachines {
		if pgM.Role == "replica" {
			if err = replaceMachine(ctx, recipe, app, pgM); err != nil {
				return err
			}
		}
	}

	// Trigger failover
	for _, pgM := range pgMachines {
		if pgM.Role == "leader" {
			_, err = recipe.RunOperation([]string{pgM.IP}, FAILOVER_SCRIPT)
			if err != nil {
				return err
			}

			if err = replaceMachine(ctx, recipe, app, pgM); err != nil {
				return err
			}
		}
	}

	return nil
}

func replaceMachine(ctx context.Context, recipe *recipes.Recipe, app *api.App, pgM PGMachine) error {
	stopInput := api.StopMachineInput{
		AppID: app.ID,
		ID:    pgM.Machine.ID,
		// KillTimeoutSecs: 10,
	}

	fmt.Printf("Stopping machine: %s\n", pgM.Machine.ID)

	_, err := recipe.Client.API().StopMachine(ctx, stopInput)
	if err != nil {
		return err
	}

	newConfig := pgM.Machine.Config
	newConfig["image"] = "flyio/postgres:14"

	launchInput := api.LaunchMachineInput{
		AppID:  app.ID,
		Region: pgM.Machine.Region,
		Config: &newConfig,
	}

	fmt.Printf("Launching new machine to replace %s\n", pgM.Machine.ID)

	machine, _, err := recipe.Client.API().LaunchMachine(ctx, launchInput)
	if err != nil {
		return err
	}
	fmt.Printf("Machine %s replaced with %s", pgM.Machine.ID, machine.ID)

	return nil
}

package recipes

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/internal/client"
	"github.com/superfly/flyctl/pkg/agent"
)

type Recipe struct {
	Agent  *agent.Client
	App    *api.App
	Client *client.Client
	Ctx    context.Context
	Dialer *agent.Dialer
}

func NewRecipe(ctx context.Context, app *api.App) (*Recipe, error) {
	client := client.New()

	agentclient, err := agent.Establish(ctx, client.API())
	if err != nil {
		return nil, errors.Wrap(err, "can't establish agent")
	}

	dialer, err := agentclient.Dialer(ctx, &app.Organization)
	if err != nil {
		return nil, fmt.Errorf("ssh: can't build tunnel for %s: %s\n", app.Organization.Slug, err)
	}

	return &Recipe{
		Ctx:    ctx,
		Client: client,
		Agent:  agentclient,
		Dialer: &dialer,
		App:    app,
	}, nil
}

// Helper for building tunnel
func (r *Recipe) BuildTunnel() error {
	r.Client.IO.StartProgressIndicatorMsg("Connecting to tunnel")
	if err := r.Agent.WaitForTunnel(r.Ctx, &r.App.Organization); err != nil {
		return errors.Wrapf(err, "tunnel unavailable")
	}
	r.Client.IO.StopProgressIndicator()

	return nil
}

func (r *Recipe) RunOperation(addrs []string, command string) ([]*RecipeOperation, error) {
	var operations []*RecipeOperation
	for _, addr := range addrs {
		fmt.Printf("Running %q against %s...\n", command, addr)
		op := NewRecipeOperation(r, addr, command)
		if err := op.Run(r.Ctx); err != nil {
			return nil, err
		}
		operations = append(operations, op)
		if op.ErrorMessage != "" {
			fmt.Printf("\n %s", op.ErrorMessage)
		}
	}
	return operations, nil
}

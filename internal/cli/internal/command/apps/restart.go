package apps

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sammccord/flyctl/pkg/iostreams"

	"github.com/sammccord/flyctl/internal/cli/internal/command"
	"github.com/sammccord/flyctl/internal/cli/internal/flag"
	"github.com/sammccord/flyctl/internal/client"
)

func newRestart() *cobra.Command {
	const (
		long = `The APPS RESTART command will restart all running vms. 
`
		short = "Restart an application"
		usage = "restart [APPNAME]"
	)

	restart := command.New(usage, short, long, RunRestart,
		command.RequireSession)

	restart.Args = cobra.ExactArgs(1)

	return restart
}

// TODO: make internal once the restart package is removed
func RunRestart(ctx context.Context) error {
	client := client.FromContext(ctx).API()

	appName := flag.FirstArg(ctx)
	if _, err := client.RestartApp(ctx, appName); err != nil {
		return fmt.Errorf("failed restarting app: %w", err)
	}

	io := iostreams.FromContext(ctx)
	fmt.Fprintf(io.Out, "%s is being restarted\n", appName)

	return nil
}

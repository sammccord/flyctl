package suspend

import (
	"github.com/spf13/cobra"

	"github.com/sammccord/flyctl/internal/cli/internal/command"
	"github.com/sammccord/flyctl/internal/cli/internal/command/apps"
)

// TODO: deprecate & remove
func New() *cobra.Command {
	const (
		long = `The SUSPEND command will suspend an application. 
All instances will be halted leaving the application running nowhere.
It will continue to consume networking resources (IP address). See APPS RESUME
for details on restarting it.
`
		short = "Suspend an application"
		usage = "suspend [APPNAME]"
	)

	suspend := command.New(usage, short, long, apps.RunSuspend,
		command.RequireSession)

	suspend.Args = cobra.ExactArgs(1)

	return suspend
}

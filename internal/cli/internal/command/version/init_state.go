package version

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sammccord/flyctl/internal/cli/internal/cache"
	"github.com/sammccord/flyctl/internal/cli/internal/command"
	"github.com/sammccord/flyctl/internal/cli/internal/flag"
	"github.com/sammccord/flyctl/pkg/iostreams"
)

func newInitState() *cobra.Command {
	initState := command.New(
		"init-state",
		"init-state",
		"Initialize installation state",
		runInitState)

	initState.Hidden = true

	initState.Args = cobra.ExactArgs(1)

	return initState
}

func runInitState(ctx context.Context) error {
	return executeInitState(
		iostreams.FromContext(ctx),
		cache.FromContext(ctx),
		flag.Args(ctx)[0],
	)
}

func executeInitState(io *iostreams.IOStreams, cache cache.Cache, channel string) error {
	cache.SetChannel(channel)

	fmt.Fprintf(io.ErrOut, "set channel to %s\n", channel)

	return nil
}

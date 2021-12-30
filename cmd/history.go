package cmd

import (
	"github.com/sammccord/flyctl/cmdctx"
	"github.com/sammccord/flyctl/internal/client"

	"github.com/sammccord/flyctl/docstrings"

	"github.com/sammccord/flyctl/cmd/presenters"
)

func newHistoryCommand(client *client.Client) *Command {
	historyStrings := docstrings.Get("history")
	return BuildCommand(nil, runHistory, historyStrings.Usage, historyStrings.Short, historyStrings.Long, client, requireSession, requireAppName)
}

func runHistory(commandContext *cmdctx.CmdContext) error {
	ctx := commandContext.Command.Context()

	changes, err := commandContext.Client.API().GetAppChanges(ctx, commandContext.AppName)
	if err != nil {
		return err
	}

	return commandContext.Frender(cmdctx.PresenterOption{Presentable: &presenters.AppHistory{AppChanges: changes}})
}

package cmd

import (
	"github.com/sammccord/flyctl/cmdctx"
	"github.com/sammccord/flyctl/internal/client"

	"github.com/sammccord/flyctl/docstrings"

	"github.com/sammccord/flyctl/cmd/presenters"
)

func newReleasesCommand(client *client.Client) *Command {
	releasesStrings := docstrings.Get("releases")
	cmd := BuildCommandKS(nil, runReleases, releasesStrings, client, requireSession, requireAppName)
	return cmd
}

func runReleases(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	releases, err := cmdCtx.Client.API().GetAppReleases(ctx, cmdCtx.AppName, 25)
	if err != nil {
		return err
	}
	return cmdCtx.Render(&presenters.Releases{Releases: releases})
}

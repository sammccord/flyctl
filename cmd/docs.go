package cmd

import (
	"fmt"

	"github.com/sammccord/flyctl/cmdctx"
	"github.com/sammccord/flyctl/internal/client"

	"github.com/sammccord/flyctl/docstrings"

	"github.com/skratchdot/open-golang/open"
)

func newDocsCommand(client *client.Client) *Command {
	docsStrings := docstrings.Get("docs")
	return BuildCommand(nil, runLaunchDocs, docsStrings.Usage, docsStrings.Short, docsStrings.Long, client)
}

const docsURL = "https://fly.io/docs/"

func runLaunchDocs(ctx *cmdctx.CmdContext) error {
	fmt.Println("Opening", docsURL)
	return open.Run(docsURL)
}

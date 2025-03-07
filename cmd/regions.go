package cmd

import (
	"github.com/sammccord/flyctl/cmdctx"
	"github.com/sammccord/flyctl/internal/client"

	"github.com/sammccord/flyctl/api"
	"github.com/sammccord/flyctl/docstrings"

	"github.com/spf13/cobra"
)

func newRegionsCommand(client *client.Client) *Command {
	regionsStrings := docstrings.Get("regions")

	cmd := BuildCommandKS(nil, nil, regionsStrings, client, requireAppName, requireSession)

	addStrings := docstrings.Get("regions.add")
	addCmd := BuildCommandKS(cmd, runRegionsAdd, addStrings, client, requireSession, requireAppName)
	addCmd.Args = cobra.MinimumNArgs(1)
	addCmd.AddStringFlag(StringFlagOpts{
		Name:        "group",
		Description: "The process group to add the region to",
		Default:     "",
	})

	removeStrings := docstrings.Get("regions.remove")
	removeCmd := BuildCommandKS(cmd, runRegionsRemove, removeStrings, client, requireSession, requireAppName)
	removeCmd.Args = cobra.MinimumNArgs(1)
	removeCmd.AddStringFlag(StringFlagOpts{
		Name:        "group",
		Description: "The process group to remove the region from",
		Default:     "",
	})

	setStrings := docstrings.Get("regions.set")
	setCmd := BuildCommandKS(cmd, runRegionsSet, setStrings, client, requireSession, requireAppName)
	setCmd.Args = cobra.MinimumNArgs(1)
	setCmd.AddStringFlag(StringFlagOpts{
		Name:        "group",
		Description: "The process group to set regions for",
		Default:     "",
	})

	setBackupStrings := docstrings.Get("regions.backup")
	setBackupCmd := BuildCommand(cmd, runBackupRegionsSet, setBackupStrings.Usage, setBackupStrings.Short, setBackupStrings.Long, client, requireSession, requireAppName)
	setBackupCmd.Args = cobra.MinimumNArgs(1)

	listStrings := docstrings.Get("regions.list")
	BuildCommand(cmd, runRegionsList, listStrings.Usage, listStrings.Short, listStrings.Long, client, requireSession, requireAppName)

	return cmd
}

func runRegionsAdd(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	group := cmdCtx.Config.GetString("group")
	input := api.ConfigureRegionsInput{
		AppID:        cmdCtx.AppName,
		Group:        group,
		AllowRegions: cmdCtx.Args,
	}

	regions, backupRegions, err := cmdCtx.Client.API().ConfigureRegions(ctx, input)
	if err != nil {
		return err
	}

	printRegions(cmdCtx, regions, backupRegions)

	return nil
}

func runRegionsRemove(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	group := cmdCtx.Config.GetString("group")
	input := api.ConfigureRegionsInput{
		AppID:       cmdCtx.AppName,
		Group:       group,
		DenyRegions: cmdCtx.Args,
	}

	regions, backupRegions, err := cmdCtx.Client.API().ConfigureRegions(ctx, input)
	if err != nil {
		return err
	}

	printRegions(cmdCtx, regions, backupRegions)

	return nil
}

func runRegionsSet(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	addList := make([]string, 0)
	delList := make([]string, 0)

	// Get the Region List
	regions, _, err := cmdCtx.Client.API().ListAppRegions(ctx, cmdCtx.AppName)
	if err != nil {
		return err
	}

	addList = append(addList, cmdCtx.Args...)

	for _, er := range regions {
		found := false
		for _, r := range cmdCtx.Args {
			if r == er.Code {
				found = true
				break
			}
		}
		if !found {
			delList = append(delList, er.Code)
		}
	}

	group := cmdCtx.Config.GetString("group")
	input := api.ConfigureRegionsInput{
		AppID:        cmdCtx.AppName,
		Group:        group,
		AllowRegions: addList,
		DenyRegions:  delList,
	}

	newregions, backupRegions, err := cmdCtx.Client.API().ConfigureRegions(ctx, input)
	if err != nil {
		return err
	}

	printRegions(cmdCtx, newregions, backupRegions)

	return nil
}

func runRegionsList(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	regions, backupRegions, err := cmdCtx.Client.API().ListAppRegions(ctx, cmdCtx.AppName)
	if err != nil {
		return err
	}

	printRegions(cmdCtx, regions, backupRegions)

	return nil
}

func runBackupRegionsSet(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	input := api.ConfigureRegionsInput{
		AppID:         cmdCtx.AppName,
		BackupRegions: cmdCtx.Args,
	}

	regions, backupRegions, err := cmdCtx.Client.API().ConfigureRegions(ctx, input)
	if err != nil {
		return err
	}

	printRegions(cmdCtx, regions, backupRegions)

	return nil
}

func printRegions(ctx *cmdctx.CmdContext, regions []api.Region, backupRegions []api.Region) {

	if ctx.OutputJSON() {
		ctx.WriteJSON(regions)
		return
	}

	verbose := ctx.GlobalConfig.GetBool("verbose")

	if verbose {
		ctx.Status("regions", cmdctx.STITLE, "Current Region Pool:")
	} else {
		ctx.Status("regions", cmdctx.STITLE, "Region Pool: ")
	}

	for _, r := range regions {
		if verbose {
			ctx.Statusf("regions", cmdctx.SINFO, "  %s  %s\n", r.Code, r.Name)
		} else {
			ctx.Status("regions", cmdctx.SINFO, r.Code)
		}
	}

	if verbose {
		ctx.Status("backupRegions", cmdctx.STITLE, "Current Backup Region Pool:")
	} else {
		ctx.Status("backupRegions", cmdctx.STITLE, "Backup Region: ")
	}

	for _, r := range backupRegions {
		if verbose {
			ctx.Statusf("backupRegions", cmdctx.SINFO, "  %s  %s\n", r.Code, r.Name)
		} else {
			ctx.Status("backupRegions", cmdctx.SINFO, r.Code)
		}
	}
}

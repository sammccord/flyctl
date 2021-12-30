package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/inancgumus/screen"
	"github.com/sammccord/flyctl/cmdctx"
	"github.com/sammccord/flyctl/internal/client"

	"github.com/sammccord/flyctl/api"
	"github.com/sammccord/flyctl/docstrings"
	"github.com/segmentio/textio"
	"github.com/spf13/cobra"

	"github.com/logrusorgru/aurora"
	"github.com/sammccord/flyctl/cmd/presenters"
)

func newStatusCommand(client *client.Client) *Command {
	statusStrings := docstrings.Get("status")
	cmd := BuildCommandKS(nil, runStatus, statusStrings, client, requireSession, requireAppNameAsArg)

	//TODO: Move flag descriptions to docstrings
	cmd.AddBoolFlag(BoolFlagOpts{Name: "all", Description: "Show completed instances"})
	cmd.AddBoolFlag(BoolFlagOpts{Name: "deployment", Description: "Always show deployment status"})
	cmd.AddBoolFlag(BoolFlagOpts{Name: "watch", Description: "Refresh details"})
	cmd.AddIntFlag(IntFlagOpts{Name: "rate", Description: "Refresh Rate for --watch", Default: 5})
	cmd.Command.Flags().String("wtf", "defaultwtf", "wtf usage")

	// cmd.Command.Flag()

	allocStatusStrings := docstrings.Get("status.instance")
	allocStatusCmd := BuildCommand(cmd, runAllocStatus, allocStatusStrings.Usage, allocStatusStrings.Short, allocStatusStrings.Long, client, requireSession, requireAppName)
	allocStatusCmd.Args = cobra.ExactArgs(1)
	return cmd
}

func runStatus(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	watch := cmdCtx.Config.GetBool("watch")
	refreshRate := cmdCtx.Config.GetInt("rate")
	refreshCount := 1
	showDeploymentStatus := cmdCtx.Config.GetBool("deployment")

	if watch && cmdCtx.OutputJSON() {
		return fmt.Errorf("--watch and --json are not supported together")
	}

	for {
		var app *api.AppStatus
		var backupregions []api.Region
		var err error
		if watch {
			refreshCount = refreshCount - 1
			if refreshCount == 0 {
				refreshCount = refreshRate
				app, err = cmdCtx.Client.API().GetAppStatus(ctx, cmdCtx.AppName, cmdCtx.Config.GetBool("all"))

				if err != nil {
					return err
				}

				if app.Deployed {
					_, backupregions, err = cmdCtx.Client.API().ListAppRegions(ctx, cmdCtx.AppName)

					if err != nil {
						return err
					}

				}
				screen.Clear()
				screen.MoveTopLeft()
				fmt.Printf("%s %s %s\n\n", aurora.Bold(app.Name), aurora.Italic("at:"), aurora.Bold(time.Now().UTC().Format("15:04:05")))
			} else {
				screen.MoveTopLeft()
				if app != nil {
					fmt.Printf("%s %s %s\n\n", aurora.Bold(app.Name), aurora.Italic("at:"), aurora.Bold(time.Now().UTC().Format("15:04:05")))
				} else {
					fmt.Printf("%s %s %s\n\n", aurora.Bold(cmdCtx.AppName), aurora.Italic("at:"), aurora.Bold(time.Now().UTC().Format("15:04:05")))
				}
				time.Sleep(time.Second)
				continue
			}
		} else {
			app, err = cmdCtx.Client.API().GetAppStatus(ctx, cmdCtx.AppName, cmdCtx.Config.GetBool("all"))

			if err != nil {
				return err
			}

			if app.Deployed {
				_, backupregions, err = cmdCtx.Client.API().ListAppRegions(ctx, cmdCtx.AppName)

				if err != nil {
					return err
				}

			}
			if err != nil {
				return err
			}
		}

		err = cmdCtx.Frender(cmdctx.PresenterOption{Presentable: &presenters.AppStatus{AppStatus: *app}, HideHeader: true, Vertical: true, Title: "App"})
		if err != nil {
			return err
		}

		// If JSON output, everything has been printed, so return
		if !watch && cmdCtx.OutputJSON() {
			return nil
		}

		// Continue formatted output
		if !app.Deployed {
			fmt.Println(`App has not been deployed yet.`)
			// exit if not watching, stay looping if we are
			if !watch {
				return nil
			}
		}

		if app.DeploymentStatus != nil {
			if (app.DeploymentStatus.Version == app.Version && app.DeploymentStatus.Status != "cancelled") || showDeploymentStatus {

				err = cmdCtx.Frender(cmdctx.PresenterOption{
					Presentable: &presenters.DeploymentStatus{Status: app.DeploymentStatus},
					Vertical:    true,
					Title:       "Deployment Status",
				})

				if err != nil {
					return err
				}
			}
		}

		err = cmdCtx.Frender(cmdctx.PresenterOption{
			Presentable: &presenters.Allocations{Allocations: app.Allocations, BackupRegions: backupregions},
			Title:       "Instances",
		})

		if err != nil {
			return err
		}

		if !watch {
			return nil
		}
	}

}

func runAllocStatus(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	alloc, err := cmdCtx.Client.API().GetAllocationStatus(ctx, cmdCtx.AppName, cmdCtx.Args[0], 25)
	if err != nil {
		return err
	}

	if alloc == nil {
		return api.ErrNotFound
	}

	err = cmdCtx.Frender(
		cmdctx.PresenterOption{
			Title: "Instance",
			Presentable: &presenters.Allocations{
				Allocations: []*api.AllocationStatus{alloc},
			},
			Vertical: true,
		},
		cmdctx.PresenterOption{
			Title: "Recent Events",
			Presentable: &presenters.AllocationEvents{
				Events: alloc.Events,
			},
		},
		cmdctx.PresenterOption{
			Title: "Checks",
			Presentable: &presenters.AllocationChecks{
				Checks: alloc.Checks,
			},
		},
	)
	if err != nil {
		return err
	}

	var p io.Writer
	var pw *textio.PrefixWriter

	if !cmdCtx.OutputJSON() {
		fmt.Println(aurora.Bold("Recent Logs"))
		pw = textio.NewPrefixWriter(cmdCtx.Out, "  ")
		p = pw
	} else {
		p = cmdCtx.Out
	}

	// logPresenter := presenters.LogPresenter{HideAllocID: true, HideRegion: true, RemoveNewlines: true}
	// logPresenter.FPrint(p, ctx.OutputJSON(), alloc.RecentLogs)

	if p != cmdCtx.Out {
		_ = pw.Flush()
	}

	return nil
}

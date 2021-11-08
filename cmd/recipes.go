package cmd

import (
	"fmt"

	"github.com/superfly/flyctl/cmdctx"
	"github.com/superfly/flyctl/recipes"
)

func runRollingRebootRecipe(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()
	client := cmdCtx.Client.API()

	app, err := client.GetApp(cmdCtx.AppName)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	if err = recipes.PostgresRebootRecipe(ctx, app); err != nil {
		return err
	}

	return nil
}

func runProvisionPostgresRecipe(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()
	appName := cmdCtx.Config.GetString("name")
	if appName == "" {
		n, err := inputAppName("", false)
		if err != nil {
			return err
		}
		appName = n
	}

	orgSlug := cmdCtx.Config.GetString("organization")
	org, err := selectOrganization(ctx, cmdCtx.Client.API(), orgSlug, nil)
	if err != nil {
		return err
	}

	regionCode := cmdCtx.Config.GetString("region")
	region, err := selectRegion(ctx, cmdCtx.Client.API(), regionCode)
	if err != nil {
		return err
	}

	consulUrl := cmdCtx.Config.GetString("consul-url")
	etcdUrl := cmdCtx.Config.GetString("etcd-url")

	if consulUrl != "" && etcdUrl != "" {
		return fmt.Errorf("consulUrl and etcdUrl may not both be specified.")
	}

	volumeSize := cmdCtx.Config.GetInt("volume-size")
	if volumeSize == 0 {
		s, err := volumeSizeInput(10)
		if err != nil {
			return err
		}
		volumeSize = s
	}

	count := cmdCtx.Config.GetInt("count")
	password := cmdCtx.Config.GetString("password")
	imageRef := cmdCtx.Config.GetString("image-ref")

	p := recipes.NewPostgresProvisionRecipe(cmdCtx, recipes.PostgresProvisionConfig{
		AppName:      appName,
		Count:        count,
		ImageRef:     imageRef,
		Organization: org,
		Password:     password,
		Region:       region.Code,
		VolumeSize:   volumeSize,
		ConsulUrl:    consulUrl,
		EtcdUrl:      etcdUrl,
	})

	return p.Start()
}

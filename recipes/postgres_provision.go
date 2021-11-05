package recipes

import (
	"context"
	"fmt"

	"github.com/superfly/flyctl/api"
	"github.com/superfly/flyctl/cmdctx"
	"github.com/superfly/flyctl/flyctl"
)

type PostgresProvision struct {
	Config PostgresProvisionConfig
	cmdCtx *cmdctx.CmdContext
}

type PostgresProvisionConfig struct {
	AppName      string
	ConsulUrl    string
	Count        int
	EtcdUrl      string
	ImageRef     string
	Organization *api.Organization
	Password     string
	Region       string
	VolumeSize   int
}

func NewPostgresProvision(cmdCtx *cmdctx.CmdContext, config PostgresProvisionConfig) *PostgresProvision {
	return &PostgresProvision{cmdCtx: cmdCtx, Config: config}
}

func (p *PostgresProvision) Start() error {
	ctx := p.cmdCtx.Command.Context()
	app, err := p.createApp()
	if err != nil {
		return err
	}

	secrets, err := p.setSecrets(ctx)
	if err != nil {
		return err
	}

	for i := 0; i < p.Config.Count; i++ {
		fmt.Printf("Provisioning %d of %d machines\n", i+1, p.Config.Count)

		machineConf := p.configurePostgres()

		launchInput := api.LaunchMachineInput{
			AppID:  app.ID,
			Region: p.Config.Region,
			Config: &machineConf,
		}

		machine, _, err := p.ctx.Client.API().LaunchMachine(p.ctx.Command.Context(), launchInput)
		if err != nil {
			return err
		}

		if err = WaitForMachineState(ctx, p.cmdCtx.Client, p.Config.AppName, machine.ID, "started"); err != nil {
			return err
		}
	}

	fmt.Printf("Connection string: postgres://postgres:%s@%s.internal:5432\n", secrets["OPERATOR_PASSWORD"], p.Config.AppName)
	return err
}

func (p *PostgresProvision) configurePostgres() api.MachineConfig {
	machineConfig := flyctl.NewMachineConfig()

	// Set env
	env := map[string]string{
		"PRIMARY_REGION": p.Config.Region,
	}

	machineConfig.SetEnvVariables(env)
	machineConfig.Config["size"] = "shared-cpu-1x"
	machineConfig.Config["image"] = p.Config.ImageRef
	machineConfig.Config["restart"] = map[string]string{
		"policy": "no",
	}

	// Set mounts
	mounts := make([]map[string]interface{}, 0)
	mounts = append(mounts, map[string]interface{}{
		"volume":    fmt.Sprintf("pg_data_%s", GenerateSecureToken(5)),
		"size_gb":   p.Config.VolumeSize,
		"encrypted": false,
		"path":      "/data",
	})
	machineConfig.Config["mounts"] = mounts

	return api.MachineConfig(machineConfig.Config)
}

func (p *PostgresProvision) createApp() (*api.App, error) {
	fmt.Println("Creating app...")
	appInput := api.CreateAppInput{
		OrganizationID:  p.Config.Organization.ID,
		Name:            p.Config.AppName,
		PreferredRegion: &p.Config.Region,
		// TODO - We should use constants to reference this.
		Runtime: "FIRECRACKER",
	}
<<<<<<< HEAD
	return p.ctx.Client.API().CreateApp(p.ctx.Command.Context(), appInput)
=======
	return p.cmdCtx.Client.API().CreateApp(appInput)
>>>>>>> 0808a8f... Messy, but progress
}

func (p *PostgresProvision) setSecrets(ctx context.Context) (map[string]string, error) {
	fmt.Println("Setting secrets...")

	secrets := map[string]string{
		"FLY_APP_NAME":      p.Config.AppName, // TODO - Move this to web.
		"FLY_REGION":        p.Config.Region,
		"SU_PASSWORD":       GenerateSecureToken(15),
		"REPL_PASSWORD":     GenerateSecureToken(15),
		"OPERATOR_PASSWORD": GenerateSecureToken(15),
	}
	fmt.Printf("Secrets %+v", secrets)
	if p.Config.Password != "" {
		secrets["OPERATOR_PASSWORD"] = p.Config.Password
	}
	if p.Config.ConsulUrl != "" {
		secrets["CONSUL_URL"] = p.Config.ConsulUrl
	}
	if p.Config.EtcdUrl != "" {
		secrets["ETCD_URL"] = p.Config.EtcdUrl
	}

<<<<<<< HEAD
	_, err := p.ctx.Client.API().SetSecrets(p.ctx.Command.Context(), p.Config.AppName, secrets)
=======
	_, err := p.cmdCtx.Client.API().SetSecrets(ctx, p.Config.AppName, secrets)
>>>>>>> 0808a8f... Messy, but progress

	return secrets, err
}

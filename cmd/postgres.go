package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/hashicorp/go-version"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"github.com/sammccord/flyctl/api"
	"github.com/sammccord/flyctl/cmd/presenters"
	"github.com/sammccord/flyctl/cmdctx"
	"github.com/sammccord/flyctl/docstrings"
	"github.com/sammccord/flyctl/helpers"
	"github.com/sammccord/flyctl/internal/client"
	"github.com/sammccord/flyctl/pkg/agent"
	"github.com/spf13/cobra"
)

type PostgresClusterOption struct {
	Name     string
	ImageRef string
	Count    int
}
type PostgresConfiguration struct {
	Name             string
	Description      string
	VmSize           string
	MemoryMb         int
	DiskGb           int
	ClusteringOption PostgresClusterOption
}

func postgresConfigurations() []PostgresConfiguration {
	return []PostgresConfiguration{
		{
			Description:      "Development - Single node, 1x shared CPU, 256MB RAM, 10GB disk",
			VmSize:           "shared-cpu-1x",
			MemoryMb:         256,
			DiskGb:           10,
			ClusteringOption: standalonePostgres(),
		},
		{
			Description:      "Production - Highly available, 1x shared CPU, 256MB RAM, 10GB disk",
			VmSize:           "shared-cpu-1x",
			MemoryMb:         256,
			DiskGb:           10,
			ClusteringOption: highlyAvailablePostgres(),
		},
		{
			Description:      "Production - Highly available, 1x Dedicated CPU, 2GB RAM, 50GB disk",
			VmSize:           "dedicated-cpu-1x",
			MemoryMb:         2048,
			DiskGb:           50,
			ClusteringOption: highlyAvailablePostgres(),
		},
		{
			Description:      "Production - Highly available, 2x Dedicated CPU's, 4GB RAM, 100GB disk",
			VmSize:           "dedicated-cpu-2x",
			MemoryMb:         4096,
			DiskGb:           100,
			ClusteringOption: highlyAvailablePostgres(),
		},
		{
			Description:      "Production - Highly available, 4x Dedicated CPU's, 8GB RAM, 200GB disk",
			VmSize:           "dedicated-cpu-4x",
			MemoryMb:         8192,
			DiskGb:           200,
			ClusteringOption: highlyAvailablePostgres(),
		},
		{
			Description: "Specify custom configuration",
			VmSize:      "",
			MemoryMb:    0,
			DiskGb:      0,
		},
	}
}

func standalonePostgres() PostgresClusterOption {
	return PostgresClusterOption{
		Name:     "Development (Single node)",
		ImageRef: "flyio/postgres-standalone",
		Count:    1,
	}
}

func highlyAvailablePostgres() PostgresClusterOption {
	return PostgresClusterOption{
		Name:     "Production (Highly available)",
		ImageRef: "flyio/postgres",
		Count:    2,
	}
}

func postgresClusteringOptions() []PostgresClusterOption {
	return []PostgresClusterOption{
		standalonePostgres(),
		highlyAvailablePostgres(),
	}
}

func newPostgresCommand(client *client.Client) *Command {
	domainsStrings := docstrings.Get("postgres")
	cmd := BuildCommandKS(nil, nil, domainsStrings, client, requireSession)
	cmd.Aliases = []string{"pg"}

	listStrings := docstrings.Get("postgres.list")
	listCmd := BuildCommandKS(cmd, runPostgresList, listStrings, client, requireSession)
	listCmd.Args = cobra.MaximumNArgs(1)

	createStrings := docstrings.Get("postgres.create")
	createCmd := BuildCommandKS(cmd, runCreatePostgresCluster, createStrings, client, requireSession)
	createCmd.AddStringFlag(StringFlagOpts{Name: "organization", Description: "the organization that will own the app"})
	createCmd.AddStringFlag(StringFlagOpts{Name: "name", Description: "the name of the new app"})
	createCmd.AddStringFlag(StringFlagOpts{Name: "region", Description: "the region to launch the new app in"})
	createCmd.AddStringFlag(StringFlagOpts{Name: "password", Description: "the superuser password. one will be generated for you if you leave this blank"})
	createCmd.AddStringFlag(StringFlagOpts{Name: "volume-size", Description: "the size in GB for volumes"})
	createCmd.AddStringFlag(StringFlagOpts{Name: "vm-size", Description: "the size of the VM"})
	createCmd.AddStringFlag(StringFlagOpts{Name: "availability", Description: "either 'standalone' for development or 'high' for production"})

	createCmd.AddStringFlag(StringFlagOpts{Name: "image-ref", Hidden: true})
	createCmd.AddStringFlag(StringFlagOpts{Name: "snapshot-id", Description: "Creates the volume with the contents of the snapshot"})

	connectStrings := docstrings.Get("postgres.connect")
	connectCmd := BuildCommandKS(cmd, runPostgresConnect, connectStrings, client, requireSession, requireAppNameAsArg)
	connectCmd.AddStringFlag(StringFlagOpts{Name: "database", Description: "The postgres database to connect to"})
	connectCmd.AddStringFlag(StringFlagOpts{Name: "user", Description: "The postgres user to connect with"})
	connectCmd.AddStringFlag(StringFlagOpts{Name: "password", Description: "The postgres user password"})

	attachStrngs := docstrings.Get("postgres.attach")
	attachCmd := BuildCommandKS(cmd, runAttachPostgresCluster, attachStrngs, client, requireSession, requireAppName)
	attachCmd.AddStringFlag(StringFlagOpts{Name: "postgres-app", Description: "the postgres cluster to attach to the app"})
	attachCmd.AddStringFlag(StringFlagOpts{Name: "database-name", Description: "database to use, defaults to a new database with the same name as the app"})
	attachCmd.AddStringFlag(StringFlagOpts{Name: "database-user", Description: "the database user to create, defaults to creating a user with the same name as the consuming app"})
	attachCmd.AddStringFlag(StringFlagOpts{Name: "variable-name", Description: "the env variable name that will be added to the app. Defaults to DATABASE_URL"})

	detachStrngs := docstrings.Get("postgres.detach")
	detachCmd := BuildCommandKS(cmd, runDetachPostgresCluster, detachStrngs, client, requireSession, requireAppName)
	detachCmd.AddStringFlag(StringFlagOpts{Name: "postgres-app", Description: "the postgres cluster to detach from the app"})

	dbStrings := docstrings.Get("postgres.db")
	dbCmd := BuildCommandKS(cmd, nil, dbStrings, client, requireSession)

	listDBStrings := docstrings.Get("postgres.db.list")
	listDBCmd := BuildCommandKS(dbCmd, runListPostgresDatabases, listDBStrings, client, requireSession, requireAppNameAsArg)
	listDBCmd.Args = cobra.ExactArgs(1)

	usersStrings := docstrings.Get("postgres.users")
	usersCmd := BuildCommandKS(cmd, nil, usersStrings, client, requireSession)

	usersListStrings := docstrings.Get("postgres.users.list")
	usersListCmd := BuildCommandKS(usersCmd, runListPostgresUsers, usersListStrings, client, requireSession, requireAppNameAsArg)
	usersListCmd.Args = cobra.ExactArgs(1)

	return cmd
}

func runPostgresList(ctx *cmdctx.CmdContext) error {
	apps, err := ctx.Client.API().GetApps(context.Background(), api.StringPointer("postgres_cluster"))
	if err != nil {
		return err
	}

	if ctx.OutputJSON() {
		ctx.WriteJSON(apps)
		return nil
	}

	return ctx.Render(&presenters.Apps{Apps: apps})
}

func runCreatePostgresCluster(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	name := cmdCtx.Config.GetString("name")
	if name == "" {
		n, err := inputAppName("", false)
		if err != nil {
			return err
		}
		name = n
	}

	orgSlug := cmdCtx.Config.GetString("organization")
	org, err := selectOrganization(ctx, cmdCtx.Client.API(), orgSlug, nil)
	if err != nil {
		return err
	}

	regionCode := cmdCtx.Config.GetString("region")
	var region *api.Region
	region, err = selectRegion(ctx, cmdCtx.Client.API(), regionCode)
	if err != nil {
		return err
	}

	input := api.CreatePostgresClusterInput{
		OrganizationID: org.ID,
		Name:           name,
		Region:         api.StringPointer(region.Code),
	}

	volumeSize := cmdCtx.Config.GetInt("volume-size")
	vmSizeName := cmdCtx.Config.GetString("vm-size")
	availability := cmdCtx.Config.GetString("availability")

	customConfig := volumeSize != 0 || vmSizeName != "" || availability != ""

	var pgConfig *PostgresConfiguration
	var vmSize *api.VMSize

	// If no custom configuration settings have been passed in, prompt user to select
	// from a list of pre-defined configurations or opt into specifying a custom
	// configuration.
	if !customConfig {
		selectedCfg := 0
		options := []string{}
		for _, cfg := range postgresConfigurations() {
			options = append(options, cfg.Description)
		}
		prompt := &survey.Select{
			Message:  "Select configuration:",
			Options:  options,
			PageSize: len(postgresConfigurations()),
		}
		if err := survey.AskOne(prompt, &selectedCfg); err != nil {
			return err
		}
		pgConfig = &postgresConfigurations()[selectedCfg]

		if pgConfig.VmSize == "" {
			// User has opted into choosing a custom configuration.
			customConfig = true
		}
	}

	if customConfig {
		option, err := postgresClusteringOptionsInput(availability)
		if err != nil {
			return err
		}

		input.Count = &option.Count
		input.ImageRef = &option.ImageRef

		// Resolve VM size
		vmSize, err = selectVMSize(ctx, cmdCtx.Client.API(), vmSizeName)
		if err != nil {
			return err
		}
		input.VMSize = api.StringPointer(vmSize.Name)

		// Resolve volume size
		if volumeSize == 0 {
			volumeSize, err = volumeSizeInput(10)
			if err != nil {
				return err
			}
		}
		input.VolumeSizeGB = api.IntPointer(volumeSize)

	} else {
		// Resolve configuration from pre-defined configuration.
		vmSize, err = selectVMSize(ctx, cmdCtx.Client.API(), pgConfig.VmSize)
		if err != nil {
			return err
		}
		input.VMSize = api.StringPointer(vmSize.Name)
		input.VolumeSizeGB = api.IntPointer(pgConfig.DiskGb)

		input.Count = api.IntPointer(pgConfig.ClusteringOption.Count)

		if imageRef := cmdCtx.Config.GetString("image-ref"); imageRef != "" {
			input.ImageRef = api.StringPointer(imageRef)
		} else {
			input.ImageRef = &pgConfig.ClusteringOption.ImageRef
		}
	}

	if password := cmdCtx.Config.GetString("password"); password != "" {
		input.Password = api.StringPointer(password)
	}

	snapshot := cmdCtx.Config.GetString("snapshot-id")
	if snapshot != "" {
		input.SnapshotID = api.StringPointer(snapshot)
	}

	fmt.Fprintf(cmdCtx.Out, "Creating postgres cluster %s in organization %s\n", name, org.Slug)

	_, err = runApiCreatePostgresCluster(cmdCtx, org.Slug, &input)

	return err
}

func runApiCreatePostgresCluster(cmdCtx *cmdctx.CmdContext, org string, input *api.CreatePostgresClusterInput) (*api.CreatePostgresClusterPayload, error) {
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Writer = os.Stderr
	s.Prefix = "Launching..."
	s.Start()

	payload, err := cmdCtx.Client.API().CreatePostgresCluster(cmdCtx.Command.Context(), *input)
	if err != nil {
		return nil, err
	}

	s.FinalMSG = fmt.Sprintf("Postgres cluster %s created\n", payload.App.Name)
	s.Stop()

	fmt.Printf("  Username:    %s\n", payload.Username)
	fmt.Printf("  Password:    %s\n", payload.Password)
	fmt.Printf("  Hostname:    %s.internal\n", payload.App.Name)
	fmt.Printf("  Proxy Port:  5432\n")
	fmt.Printf("  PG Port: 5433\n")

	fmt.Println(aurora.Italic("Save your credentials in a secure place, you won't be able to see them again!"))
	fmt.Println()

	cancelCtx := cmdCtx.Command.Context()
	cmdCtx.AppName = payload.App.Name
	err = watchDeployment(cancelCtx, cmdCtx)

	if isCancelledError(err) {
		err = nil
	}

	if err == nil {
		fmt.Println()
		fmt.Println(aurora.Bold("Connect to postgres"))
		fmt.Printf("Any app within the %s organization can connect to postgres using the above credentials and the hostname \"%s.internal.\"\n", org, payload.App.Name)
		fmt.Printf("For example: postgres://%s:%s@%s.internal:%d\n", payload.Username, payload.Password, payload.App.Name, 5432)

		fmt.Println()
		fmt.Println("See the postgres docs for more information on next steps, managing postgres, connecting from outside fly:  https://fly.io/docs/reference/postgres/")
	}

	return payload, err
}

func runAttachPostgresCluster(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	postgresAppName := cmdCtx.Config.GetString("postgres-app")
	appName := cmdCtx.AppName

	dbName := cmdCtx.Config.GetString("database-name")
	if dbName == "" {
		dbName = appName
	}
	dbName = strings.ToLower(strings.ReplaceAll(dbName, "-", "_"))

	varName := cmdCtx.Config.GetString("variable-name")
	if varName == "" {
		varName = "DATABASE_URL"
	}

	dbUser := cmdCtx.Config.GetString("database-user")
	if dbUser == "" {
		dbUser = appName
	}
	dbUser = strings.ToLower(strings.ReplaceAll(dbUser, "-", "_"))

	input := api.AttachPostgresClusterInput{
		AppID:                appName,
		PostgresClusterAppID: postgresAppName,
		ManualEntry:          true,
		DatabaseName:         api.StringPointer(dbName),
		DatabaseUser:         api.StringPointer(dbUser),
		VariableName:         api.StringPointer(varName),
	}

	client := cmdCtx.Client.API()

	app, err := client.GetApp(ctx, appName)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	pgApp, err := client.GetApp(ctx, postgresAppName)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	agentclient, err := agent.Establish(ctx, cmdCtx.Client.API())
	if err != nil {
		return errors.Wrap(err, "can't establish agent")
	}

	dialer, err := agentclient.Dialer(ctx, &pgApp.Organization)
	if err != nil {
		return fmt.Errorf("ssh: can't build tunnel for %s: %s\n", app.Organization.Slug, err)
	}

	pgCmd := NewPostgresCmd(cmdCtx, pgApp, dialer)

	secrets, err := client.GetAppSecrets(ctx, appName)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		if secret.Name == *input.VariableName {
			return fmt.Errorf("Consumer app %q already contains a secret named %s.", appName, *input.VariableName)
		}
	}
	// Check to see if database exists
	dbExists, err := pgCmd.DbExists(*input.DatabaseName)
	if err != nil {
		return err
	}
	if dbExists {
		confirm := false
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Database %q already exists. Continue with the attachment process?", *input.DatabaseName),
		}
		err = survey.AskOne(prompt, &confirm)
		if err != nil {
			return err
		}

		if !confirm {
			return nil
		}
	}

	// Check to see if user exists
	usrExists, err := pgCmd.UserExists(*input.DatabaseUser)
	if err != nil {
		return err
	}
	if usrExists {
		return fmt.Errorf("Database user %q already exists. Please specify a new database user via --database-user", *input.DatabaseUser)
	}

	// Create attachment
	_, err = client.AttachPostgresCluster(ctx, input)
	if err != nil {
		return err
	}

	// Create database if it doesn't already exist
	if !dbExists {
		dbResp, err := pgCmd.CreateDatabase(*input.DatabaseName)
		if err != nil {
			return err
		}
		if dbResp.Error != "" {
			return errors.Wrap(fmt.Errorf(dbResp.Error), "executing database-create")
		}
	}

	// Create user
	pwd, err := helpers.RandString(15)
	if err != nil {
		return err
	}

	usrResp, err := pgCmd.CreateUser(*input.DatabaseUser, pwd)
	if err != nil {
		return err
	}
	if usrResp.Error != "" {
		return errors.Wrap(fmt.Errorf(usrResp.Error), "executing create-user")
	}

	// Grant access
	gaResp, err := pgCmd.GrantAccess(*input.DatabaseName, *input.DatabaseUser)
	if err != nil {
		return err
	}
	if gaResp.Error != "" {
		return errors.Wrap(fmt.Errorf(usrResp.Error), "executing grant-access")
	}

	connectionString := fmt.Sprintf("postgres://%s:%s@top2.nearest.of.%s.internal:5432/%s", *input.DatabaseUser, pwd, pgApp.Name, *input.DatabaseName)
	s := map[string]string{}
	s[*input.VariableName] = connectionString

	_, err = client.SetSecrets(ctx, appName, s)
	if err != nil {
		return err
	}

	fmt.Printf("\nPostgres cluster %s is now attached to %s\n", pgApp.Name, app.Name)
	fmt.Printf("The following secret was added to %s:\n  %s=%s\n", app.Name, *input.VariableName, connectionString)

	return nil
}

func runDetachPostgresCluster(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	postgresAppName := cmdCtx.Config.GetString("postgres-app")
	appName := cmdCtx.AppName

	client := cmdCtx.Client.API()

	app, err := client.GetApp(ctx, appName)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	pgApp, err := client.GetApp(ctx, postgresAppName)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	attachments, err := client.ListPostgresClusterAttachments(ctx, app.ID, pgApp.ID)
	if err != nil {
		return err
	}

	if len(attachments) == 0 {
		return fmt.Errorf("No attachments found")
	}

	selected := 0
	options := []string{}
	for _, opt := range attachments {
		str := fmt.Sprintf("PG Database: %s, PG User: %s, Environment variable: %s", opt.DatabaseName, opt.DatabaseUser, opt.EnvironmentVariableName)
		options = append(options, str)
	}
	prompt := &survey.Select{
		Message:  "Select the attachment that you would like to detach: ( Note: Database will not be removed! )",
		Options:  options,
		PageSize: len(attachments),
	}
	if err := survey.AskOne(prompt, &selected); err != nil {
		return err
	}

	targetAttachment := attachments[selected]

	agentclient, err := agent.Establish(ctx, client)
	if err != nil {
		return errors.Wrap(err, "can't establish agent")
	}

	dialer, err := agentclient.Dialer(ctx, &pgApp.Organization)
	if err != nil {
		return fmt.Errorf("ssh: can't build tunnel for %s: %s\n", app.Organization.Slug, err)
	}

	pgCmd := NewPostgresCmd(cmdCtx, pgApp, dialer)

	// Remove user if exists
	exists, err := pgCmd.UserExists(targetAttachment.DatabaseUser)
	if err != nil {
		return err
	}
	if exists {
		// Revoke access to suer
		raResp, err := pgCmd.RevokeAccess(targetAttachment.DatabaseName, targetAttachment.DatabaseUser)
		if err != nil {
			return err
		}
		if raResp.Error != "" {
			return errors.Wrap(fmt.Errorf(raResp.Error), "executing revoke-access")
		}

		ruResp, err := pgCmd.DeleteUser(targetAttachment.DatabaseUser)
		if err != nil {
			return err
		}
		if ruResp.Error != "" {
			return errors.Wrap(fmt.Errorf(ruResp.Error), "executing user-delete")
		}
	}

	// Remove secret from consumer app.
	_, err = client.UnsetSecrets(ctx, appName, []string{targetAttachment.EnvironmentVariableName})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("Secret %q was scheduled to be removed from app %s\n", targetAttachment.EnvironmentVariableName, app.Name)
	}

	input := api.DetachPostgresClusterInput{
		AppID:                       appName,
		PostgresClusterId:           postgresAppName,
		PostgresClusterAttachmentId: targetAttachment.ID,
	}

	if err = client.DetachPostgresCluster(ctx, input); err != nil {
		return err
	}

	fmt.Println("Detach completed successfully!")

	return nil
}

func runPostgresConnect(cmdCtx *cmdctx.CmdContext) error {
	// Minimum image version requirements
	MinPostgresStandaloneVersion := "0.0.4"
	MinPostgresHaVersion := "0.0.9"

	ctx := cmdCtx.Command.Context()
	client := cmdCtx.Client.API()

	app, err := client.GetApp(ctx, cmdCtx.AppName)
	if err != nil {
		return fmt.Errorf("get app: %w", err)
	}

	// Validate image version to ensure it's compatible with this feature.
	imageVersionStr := app.ImageDetails.Version[1:]
	imageVersion, err := version.NewVersion(imageVersionStr)
	if err != nil {
		return err
	}

	// Specify compatible versions per repo.
	requiredVersion := &version.Version{}
	if app.ImageDetails.Repository == "flyio/postgres-standalone" {
		// https://github.com/fly-apps/postgres-standalone/releases/tag/v0.0.4
		requiredVersion, err = version.NewVersion(MinPostgresStandaloneVersion)
		if err != nil {
			return err
		}
	}
	if app.ImageDetails.Repository == "flyio/postgres" {
		// https://github.com/fly-apps/postgres-ha/releases/tag/v0.0.9
		requiredVersion, err = version.NewVersion(MinPostgresHaVersion)
		if err != nil {
			return err
		}
	}

	if requiredVersion == nil {
		return fmt.Errorf("Unable to resolve image version...")
	}

	if imageVersion.LessThan(requiredVersion) {
		return fmt.Errorf(
			"Image version is not compatible. (Current: %s, Required: >= %s)\n"+
				"Please run 'flyctl image show' and update to the latest available version.",
			imageVersion, requiredVersion.String())
	}

	agentclient, err := agent.Establish(ctx, cmdCtx.Client.API())
	if err != nil {
		return errors.Wrap(err, "can't establish agent")
	}

	dialer, err := agentclient.Dialer(ctx, &app.Organization)
	if err != nil {
		return fmt.Errorf("ssh: can't build tunnel for %s: %s\n", app.Organization.Slug, err)
	}

	database := cmdCtx.Config.GetString("database")
	if database == "" {
		database = "postgres"
	}

	user := cmdCtx.Config.GetString("user")
	if user == "" {
		user = "postgres"
	}

	password := cmdCtx.Config.GetString("password")

	addr := fmt.Sprintf("%s.internal", cmdCtx.AppName)
	cmd := fmt.Sprintf("connect %s %s %s", database, user, password)

	return sshConnect(&SSHParams{
		Ctx:    cmdCtx,
		Org:    &app.Organization,
		Dialer: dialer,
		App:    cmdCtx.AppName,
		Cmd:    cmd,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}, addr)
}

func runListPostgresDatabases(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	databases, err := cmdCtx.Client.API().ListPostgresDatabases(ctx, cmdCtx.AppName)
	if err != nil {
		return err
	}

	if cmdCtx.OutputJSON() {
		cmdCtx.WriteJSON(databases)
		return nil
	}

	table := helpers.MakeSimpleTable(cmdCtx.Out, []string{"Name", "Users"})

	for _, database := range databases {
		table.Append([]string{database.Name, strings.Join(database.Users, ",")})
	}

	table.Render()

	return nil
}

func runListPostgresUsers(cmdCtx *cmdctx.CmdContext) error {
	ctx := cmdCtx.Command.Context()

	users, err := cmdCtx.Client.API().ListPostgresUsers(ctx, cmdCtx.AppName)
	if err != nil {
		return err
	}

	if cmdCtx.OutputJSON() {
		cmdCtx.WriteJSON(users)
		return nil
	}

	table := helpers.MakeSimpleTable(cmdCtx.Out, []string{"Username", "Superuser", "Databases"})

	for _, user := range users {
		table.Append([]string{user.Username, strconv.FormatBool(user.IsSuperuser), strings.Join(user.Databases, ",")})
	}

	table.Render()

	return nil
}

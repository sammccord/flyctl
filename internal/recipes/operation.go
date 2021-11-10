package recipes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/docker/docker/pkg/ioutils"
	"github.com/superfly/flyctl/api"
)

type RecipeOperation struct {
	Recipe       *Recipe
	Machine      *api.Machine
	Command      string
	Result       string
	ErrorMessage string
}

func NewRecipeOperation(recipe *Recipe, machine *api.Machine, command string) *RecipeOperation {
	return &RecipeOperation{Machine: machine, Command: command, Recipe: recipe}
}

func (o *RecipeOperation) RunHTTPCommand(ctx context.Context, method, endpoint string) error {
	baseUri := fmt.Sprintf("http://%s:%s@[%s]:4280", o.Recipe.App.Name, o.Recipe.AuthToken, o.MachineIP())
	targetEndpoint := fmt.Sprintf("%s/%s", baseUri, endpoint)

	req, err := http.NewRequest(method, targetEndpoint, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// TODO - Parse response to resolve any errors.
	o.Result = string(b)

	return nil
}

func (o *RecipeOperation) RunSSHCommand(ctx context.Context) error {
	var inBuf bytes.Buffer
	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	stdoutWriter := ioutils.NewWriteCloserWrapper(&outBuf, func() error { return nil })
	stderrWriter := ioutils.NewWriteCloserWrapper(&errBuf, func() error { return nil })
	inReader := ioutils.NewReadCloserWrapper(&inBuf, func() error { return nil })

	formattedAddr := fmt.Sprintf("[%s]", o.Addr())

	err := sshConnect(&SSHParams{
		Ctx:       ctx,
		Org:       &o.Recipe.App.Organization,
		Dialer:    *o.Recipe.Dialer,
		ApiClient: o.Recipe.Client.API(),
		App:       o.Recipe.App.Name,
		Cmd:       o.Command,
		Stdin:     inReader,
		Stdout:    stdoutWriter,
		Stderr:    stderrWriter,
	}, formattedAddr)

	if err != nil {
		o.ErrorMessage = err.Error()
		return err
	}

	o.Result = strings.TrimSuffix(outBuf.String(), "\r\n")
	o.Result = strings.Trim(o.Result, "\"")
	o.ErrorMessage = errBuf.String()

	return nil
}

func (o *RecipeOperation) Addr() string {
	return o.Machine.IPs.Nodes[0].IP
}

func (o *RecipeOperation) MachineIP() string {
	peerIP := net.ParseIP(o.Addr())
	var natsIPBytes [16]byte
	copy(natsIPBytes[0:], peerIP[0:6])
	natsIPBytes[15] = 3

	return net.IP(natsIPBytes[:]).String()
}

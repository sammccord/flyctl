package recipes

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/pkg/ioutils"
)

type RecipeOperation struct {
	Recipe       *Recipe
	Addr         string
	Command      string
	Result       string
	ErrorMessage string
}

func NewRecipeOperation(recipe *Recipe, addr, command string) *RecipeOperation {
	return &RecipeOperation{Addr: addr, Command: command, Recipe: recipe}
}

func (o *RecipeOperation) Run(ctx context.Context) error {
	var inBuf bytes.Buffer
	var errBuf bytes.Buffer
	var outBuf bytes.Buffer
	stdoutWriter := ioutils.NewWriteCloserWrapper(&outBuf, func() error { return nil })
	stderrWriter := ioutils.NewWriteCloserWrapper(&errBuf, func() error { return nil })
	inReader := ioutils.NewReadCloserWrapper(&inBuf, func() error { return nil })

	formattedAddr := fmt.Sprintf("[%s]", o.Addr)

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

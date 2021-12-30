//go:build windows
// +build windows

package agent

import (
	"context"
	"fmt"

	"github.com/sammccord/flyctl/api"
)

func StartDaemon(ctx context.Context, api *api.Client, cmd string) (*Client, error) {
	return nil, fmt.Errorf("can't start agent on this platform (this is a bug, please report)")
}

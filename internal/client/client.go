package client

import (
	"errors"

	"github.com/sammccord/flyctl/api"
	"github.com/sammccord/flyctl/flyctl"
	"github.com/sammccord/flyctl/internal/buildinfo"
	"github.com/sammccord/flyctl/pkg/iostreams"
	"github.com/sammccord/flyctl/terminal"
)

var ErrNoAuthToken = errors.New("No access token available. Please login with 'flyctl auth login'")

func New() *Client {
	client := &Client{
		IO: iostreams.System(),
	}

	client.InitApi()

	return client
}

type Client struct {
	IO *iostreams.IOStreams // TODO: remove

	api *api.Client
}

func (c *Client) API() *api.Client {
	return c.api
}

func (c *Client) Authenticated() bool {
	return c.api != nil
}

func (c *Client) InitApi() bool {
	apiToken := flyctl.GetAPIToken()
	if apiToken != "" {
		apiClient := api.NewClient(apiToken, buildinfo.Name(), buildinfo.Version().String(), terminal.DefaultLogger)
		c.api = apiClient
	}
	return c.Authenticated()
}

func FromToken(token string) *Client {
	ac := api.NewClient(token, buildinfo.Name(), buildinfo.Version().String(), terminal.DefaultLogger)

	return &Client{
		api: ac,
	}
}

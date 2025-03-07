package imgsrc

import (
	"context"
	"errors"
	"fmt"

	"github.com/sammccord/flyctl/api"
	"github.com/sammccord/flyctl/flyctl"
	"github.com/sammccord/flyctl/pkg/iostreams"
	"github.com/sammccord/flyctl/terminal"
)

type ImageOptions struct {
	AppName        string
	WorkingDir     string
	DockerfilePath string
	ImageRef       string
	AppConfig      *flyctl.AppConfig
	ExtraBuildArgs map[string]string
	ImageLabel     string
	Publish        bool
	Tag            string
	Target         string
	NoCache        bool
}

type RefOptions struct {
	AppName    string
	WorkingDir string
	ImageRef   string
	AppConfig  *flyctl.AppConfig
	ImageLabel string
	Publish    bool
	Tag        string
}

type DeploymentImage struct {
	ID   string
	Tag  string
	Size int64
}

type Resolver struct {
	dockerFactory *dockerClientFactory
	apiClient     *api.Client
}

// ResolveReference returns an Image give an reference using either the local docker daemon or remote registry
func (r *Resolver) ResolveReference(ctx context.Context, streams *iostreams.IOStreams, opts RefOptions) (img *DeploymentImage, err error) {
	strategies := []imageResolver{
		&localImageResolver{},
		&remoteImageResolver{flyApi: r.apiClient},
	}

	for _, s := range strategies {
		terminal.Debugf("Trying '%s' strategy\n", s.Name())
		img, err = s.Run(ctx, r.dockerFactory, streams, opts)
		terminal.Debugf("result image:%+v error:%v\n", img, err)
		if err != nil {
			return nil, err
		}
		if img != nil {
			return img, nil
		}
	}

	return nil, fmt.Errorf("could not find image \"%s\"", opts.ImageRef)
}

// BuildImage converts source code to an image using a Dockerfile, buildpacks, or builtins.
func (r *Resolver) BuildImage(ctx context.Context, streams *iostreams.IOStreams, opts ImageOptions) (img *DeploymentImage, err error) {
	if !r.dockerFactory.mode.IsAvailable() {
		return nil, errors.New("docker is unavailable to build the deployment image")
	}

	if opts.Tag == "" {
		opts.Tag = newDeploymentTag(opts.AppName, opts.ImageLabel)
	}

	strategies := []imageBuilder{
		&buildpacksBuilder{},
		&dockerfileBuilder{},
		&builtinBuilder{},
	}

	for _, s := range strategies {
		terminal.Debugf("Trying '%s' strategy\n", s.Name())
		img, err = s.Run(ctx, r.dockerFactory, streams, opts)
		terminal.Debugf("result image:%+v error:%v\n", img, err)
		if err != nil {
			return nil, err
		}
		if img != nil {
			return img, nil
		}
	}

	return nil, errors.New("app does not have a Dockerfile or buildpacks configured. See https://fly.io/docs/reference/configuration/#the-build-section")
}

func NewResolver(daemonType DockerDaemonType, apiClient *api.Client, appName string, iostreams *iostreams.IOStreams) *Resolver {
	return &Resolver{
		dockerFactory: newDockerClientFactory(daemonType, apiClient, appName, iostreams),
		apiClient:     apiClient,
	}
}

type imageBuilder interface {
	Name() string
	Run(ctx context.Context, dockerFactory *dockerClientFactory, streams *iostreams.IOStreams, opts ImageOptions) (*DeploymentImage, error)
}

type imageResolver interface {
	Name() string
	Run(ctx context.Context, dockerFactory *dockerClientFactory, streams *iostreams.IOStreams, opts RefOptions) (*DeploymentImage, error)
}

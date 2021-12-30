package imgsrc

import (
	"context"
	"fmt"

	"github.com/sammccord/flyctl/api"
	"github.com/sammccord/flyctl/pkg/iostreams"
	"github.com/sammccord/flyctl/terminal"
)

type remoteImageResolver struct {
	flyApi *api.Client
}

func (s *remoteImageResolver) Name() string {
	return "Remote Image Reference"
}

func (s *remoteImageResolver) Run(ctx context.Context, dockerFactory *dockerClientFactory, streams *iostreams.IOStreams, opts RefOptions) (*DeploymentImage, error) {
	ref := imageRefFromOpts(opts)
	if ref == "" {
		terminal.Debug("no image reference found, skipping")
		return nil, nil
	}

	fmt.Fprintf(streams.ErrOut, "Searching for image '%s' remotely...\n", ref)

	img, err := s.flyApi.ResolveImageForApp(ctx, opts.AppName, ref)
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, nil
	}

	fmt.Fprintf(streams.ErrOut, "image found: %s\n", img.ID)

	di := &DeploymentImage{
		ID:   img.ID,
		Tag:  img.Ref,
		Size: int64(img.CompressedSize),
	}

	return di, nil
}

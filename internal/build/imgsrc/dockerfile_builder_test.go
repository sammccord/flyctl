package imgsrc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sammccord/flyctl/flyctl"
	"github.com/sammccord/flyctl/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestBuildDockerfileApp(t *testing.T) {
	t.Skip()
	df := newDockerClientFactory(DockerDaemonTypeLocal, nil, "test-app", nil)

	dfStrategy := dockerfileBuilder{}
	testStreams, _, _, _ := iostreams.Test()

	wd, err := os.Getwd()
	assert.NoError(t, err)

	workingDir := filepath.Join(wd, "testdata", "dockerfile_app")
	configFilePath := filepath.Join(workingDir, "fly.toml")

	appConfig, err := flyctl.LoadAppConfig(configFilePath)
	assert.NoError(t, err)

	opts := ImageOptions{
		AppName:    "test-app",
		WorkingDir: workingDir,
		AppConfig:  appConfig,
		Tag:        "test-dockerfile-app",
	}

	img, err := dfStrategy.Run(context.TODO(), df, testStreams, opts)
	fmt.Printf("err: %#v %T", err, err)
	assert.NoError(t, err)
	assert.NotNil(t, img)

	assert.Equal(t, "test-dockerfile-app", img.Tag)
}

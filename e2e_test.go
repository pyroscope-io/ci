package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/pyroscope-io/ci/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func BuildImage(dockerfilePath string) func(env *testscript.Env) error {
	return func(env *testscript.Env) error {
		// TODO:
		//from := "examples/nodejs/jest"
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}

		fmt.Println("Building image")
		return buildImage(context.Background(), cli, dockerfilePath, "mytag")
	}
}

var proxyID string

func StartProxy(ctx context.Context, cli *docker.Client) string {
	// TODO: dirty check to not run the same proxy twice
	// which seems to happen when invoking the binary
	if proxyID != "" {
		return proxyID
	}

	cfg := &container.Config{
		Image: "qoomon/docker-host",
	}

	hc := &container.HostConfig{
		CapAdd: []string{"NET_ADMIN", "NET_RAW"},
	}

	fmt.Println("creating container")
	res, err := cli.ContainerCreate(ctx, cfg, hc, nil, nil, "docker-host")
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStart(context.Background(), res.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	// TODO: wait for that container to be ready

	return res.ID
}

// Make the pyroscope-ci binary available to the testscripts
func TestMain(m *testing.M) {
	exitCode := testscript.RunMain(m, map[string]func() int{
		"pyroscope-ci": func() int {
			err := cmd.RootCmd()
			if err != nil {
				return 1
			}

			return 0
		},
	})

	fmt.Println("exiting", time.Now())
	os.Exit(exitCode)
}

func TestE2E(t *testing.T) {
	// TODO: run containers with different
	fmt.Println("starting")
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	proxyIDRet := StartProxy(ctx, cli)
	t.Cleanup(func() {
		fmt.Println("Cleaning up")
		err = cli.ContainerRemove(ctx, proxyIDRet, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			panic(err)
		}
	})

	testscript.Run(t, testscript.Params{
		// BuildImage, export its name as an env var
		Setup: BuildImage("examples/nodejs/jest"),
		Dir:   "./examples/nodejs/jest",
	})
}

func buildImage(ctx context.Context, cli *docker.Client, path, tag string) error {
	// Let's remove the image to make sure it's properly built.
	// We rely on cache to rebuild it fast when existing.
	// Since the original container may still exist, we need to force the image deletion
	//	_, err := cli.ImageRemove(ctx, tag, types.ImageRemoveOptions{PruneChildren: true, Force: true})
	//	if err != nil && !client.IsErrNotFound(err) {
	//		return err
	//	}
	// TODO: ignore node_modules
	tar, err := archive.Tar(path, archive.Gzip)
	if err != nil {
		return err
	}
	res, err := cli.ImageBuild(ctx, tar, types.ImageBuildOptions{Tags: []string{tag}}) //		Remove:      true,
	//		ForceRemove: true,

	if err != nil {
		return err
	}
	defer res.Body.Close()
	rd := bufio.NewReader(res.Body)
	var line string
	for {
		var err error
		l, err := rd.ReadString('\n')
		if err == nil {
			line = l
			continue
		}
		if err == io.EOF {
			if !strings.Contains(line, "Successfully") {
				return fmt.Errorf("unexpected last message when building image: %s", line)
			}
			break
		}
		return err
	}
	return nil
}

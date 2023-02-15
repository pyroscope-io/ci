package main_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	cp "github.com/otiai10/copy"
	"github.com/pyroscope-io/ci/cmd"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

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

	os.Exit(exitCode)
}

func TestNodeJest(t *testing.T) {
	containerName, cleanup := StartProxy2(t)
	t.Cleanup(func() {
		cleanup()
	})

	testscript.Run(t, testscript.Params{
		Setup: func(env *testscript.Env) error {
			env.Vars = append(env.Vars, "PYROSCOPE_PROXY_ADDRESS="+containerName)

			imageName := "example-nodejs"
			err := BuildImage("./examples/nodejs/jest", imageName)(env)
			if err != nil {
				return err
			}
			env.Vars = append(env.Vars, "IMAGE_NAME="+imageName)
			return nil
		},
		//		Setup: BuildImage("examples/nodejs/jest", "example-nodejs"),
		Dir: "./examples/nodejs/jest",
	})
}

func TestGo(t *testing.T) {
	containerName, cleanup := StartProxy2(t)
	t.Cleanup(func() {
		cleanup()
	})

	testscript.Run(t, testscript.Params{
		Setup: func(env *testscript.Env) error {
			err := CopyFiles("./examples/go", env)
			if err != nil {
				return err
			}

			env.Vars = append(env.Vars, "PYROSCOPE_PROXY_ADDRESS="+containerName)

			imageName := "example-go"
			err = BuildImage("./examples/go", imageName)(env)
			if err != nil {
				return err
			}
			env.Vars = append(env.Vars, "IMAGE_NAME="+imageName)
			return nil
		},
		//		Setup: BuildImage("examples/nodejs/jest", "example-nodejs"),
		Dir: "./examples/go",
	})
}

func buildImage(ctx context.Context, cli *docker.Client, path, tag string) error {
	// Let's remove the image to make sure it's properly built.
	// We rely on cache to rebuild it fast when existing.
	// Since the original container may still exist, we need to force the image deletion
	_, err := cli.ImageRemove(ctx, tag, types.ImageRemoveOptions{PruneChildren: true, Force: true})
	if err != nil && !docker.IsErrNotFound(err) {
		return err
	}

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

// StartProxy2 starts a proxy that forward all requests to the host
// This is necessary since the `pyroscope-ci` binary runs in the host
// For more info see https://github.com/qoomon/docker-host
func StartProxy2(t *testing.T) (string, func()) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Name:       "docker-host",
		Hostname:   "docker-host",
		Image:      "qoomon/docker-host",
		CapAdd:     []string{"NET_ADMIN", "NET_RAW"},
		WaitingFor: wait.ForLog("Forwarding ports: 1-65535"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	//	time.Sleep(time.Second * 35)
	if err != nil {
		t.Fatal(err)
	}
	containerName, err := container.Name(ctx)
	if err != nil {
		t.Fatal(err)
	}
	containerName = strings.TrimPrefix(containerName, "/")

	return containerName, func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}
}

func StartProxyDeprecated(t *testing.T) (string, func()) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	containerName := "docker-host"

	// TODO: dirty check to not run the same proxy twice
	// which seems to happen when invoking the binary
	cfg := &container.Config{
		Image: "qoomon/docker-host",
	}

	hc := &container.HostConfig{
		CapAdd: []string{"NET_ADMIN", "NET_RAW"},
	}

	//	_, err := cli.ImagePull(ctx, cfg.Image, types.ImagePullOptions{})
	//	if err != nil {
	//		panic(err)
	//	}

	fmt.Println("creating container")
	res, err := cli.ContainerCreate(ctx, cfg, hc, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStart(context.Background(), res.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	// TODO: wait for that container to be ready

	return containerName, func() {
		fmt.Println("removing container")
		err := cli.ContainerRemove(ctx, res.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			panic(err)
		}
	}
}

func BuildImage(dockerfilePath string, imageName string) func(env *testscript.Env) error {
	return func(env *testscript.Env) error {
		cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}

		fmt.Println("Building image")
		return buildImage(context.Background(), cli, dockerfilePath, imageName)
	}
}

func CopyFiles(from string, env *testscript.Env) error {
	return cp.Copy(from, env.Cd)
}

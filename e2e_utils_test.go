//go:build e2e
// +build e2e

package main_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	cp "github.com/otiai10/copy"
	"github.com/rogpeppe/go-internal/testscript"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type setupFn func(env *testscript.Env) error

func Setup(funcs ...setupFn) setupFn {
	return func(env *testscript.Env) error {
		for _, fn := range funcs {
			err := fn(env)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// BuildImage builds an image and sets it as the IMAGE_NAME env var for usage in testscripts
func BuildImage(dockerfilePath string, imageName string) setupFn {
	return func(env *testscript.Env) error {
		cli, err := docker.NewClientWithOpts(docker.FromEnv, docker.WithAPIVersionNegotiation())
		if err != nil {
			return err
		}

		err = buildImage(context.Background(), cli, dockerfilePath, imageName)
		if err != nil {
			return err
		}
		env.Vars = append(env.Vars, "IMAGE_NAME="+imageName)
		return nil
	}
}

func SetProxyAddressEnvVar(containerName string) setupFn {
	return func(env *testscript.Env) error {
		env.Vars = append(env.Vars, "PYROSCOPE_PROXY_ADDRESS="+containerName)
		return nil
	}
}

func CopyFilesToCwd(from string) setupFn {
	return func(env *testscript.Env) error {
		return cp.Copy(from, env.Cd)
	}
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
	res, err := cli.ImageBuild(ctx, tar, types.ImageBuildOptions{Tags: []string{tag}})

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

// StartProxy starts a proxy that forward all requests to the host
// This is necessary since the `pyroscope-ci` binary runs in the host
// For more info see https://github.com/qoomon/docker-host
func StartProxy(t *testing.T) (string, func()) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		//		Name:       "docker-host",
		//		Hostname:   "docker-host",
		Image:      "qoomon/docker-host",
		CapAdd:     []string{"NET_ADMIN", "NET_RAW"},
		WaitingFor: wait.ForLog("Forwarding ports: 1-65535"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

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

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/pyroscope-io/ci/cmd"
	"github.com/rogpeppe/go-internal/testscript"
)

func CopyFiles(env *testscript.Env) error {
	from := "/Users/eduardo/work/pyroscope/ci/examples/nodejs/jest"
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	fmt.Println("Building image")
	return buildImage(context.Background(), cli, from, "mytag")
	//	err := cp.Copy(from, env.Cd, cp.Options{
	//		Skip: func(info os.FileInfo, src, dest string) (bool, error) {
	//			return strings.HasSuffix(src, ".git") || strings.HasSuffix(src, "node_modules"), nil
	//		},
	//	})

	//return err
}

// Make the pyroscope-ci binary available to the testscripts
func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"pyroscope-ci": func() int {
			err := cmd.RootCmd()
			if err != nil {
				return 1
			}

			return 0
		},
	}))
}

func TestE2E(t *testing.T) {
	testscript.Run(t, testscript.Params{
		// BuildImage, export its name as an env var
		Setup: CopyFiles,
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
	res, err := cli.ImageBuild(ctx, tar, types.ImageBuildOptions{Tags: []string{tag},
		Remove:      true,
		ForceRemove: true,
	})
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

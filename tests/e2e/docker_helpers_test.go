//go:build e2e_docker

package e2e

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcexec "github.com/testcontainers/testcontainers-go/exec"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

// newNetwork creates a throwaway Docker network so the Drupal container can
// reach MariaDB by the alias "mariadb" (matching the examples' connection
// host).
func newNetwork(t *testing.T, ctx context.Context) *testcontainers.DockerNetwork {
	t.Helper()
	nw, err := network.New(ctx)
	require.NoError(t, err, "create docker network")
	t.Cleanup(func() {
		if err := nw.Remove(ctx); err != nil {
			t.Logf("remove network: %v", err)
		}
	})
	return nw
}

// startDrupalContainer builds the Drupal+drush image from tests/e2e/Dockerfile
// (with the repo as build context) and starts it on the given network. The
// container idles (tail -f) so we can exec commands into it.
func startDrupalContainer(t *testing.T, ctx context.Context, nw *testcontainers.DockerNetwork) testcontainers.Container {
	t.Helper()

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			FromDockerfile: testcontainers.FromDockerfile{
				Context:    repoRoot,
				Dockerfile: "tests/e2e/drupal.dockerfile",
				KeepImage:  true,
			},
			Cmd:        []string{"tail", "-f", "/dev/null"},
			WaitingFor: wait.ForExec([]string{"drush", "--version"}).WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	}
	// Attach to the shared network with the "drupal" alias.
	require.NoError(t, network.WithNetwork([]string{"drupal"}, nw)(&req))

	c, err := testcontainers.GenericContainer(ctx, req)
	require.NoError(t, err, "build/start drupal container")
	t.Cleanup(func() { terminate(t, c) })
	return c
}

// execInContainer runs a command in the container and fails the test on a
// non-zero exit code. Working directory is /app (the Drupal root).
func execInContainer(t *testing.T, ctx context.Context, c testcontainers.Container, cmd ...string) {
	t.Helper()
	code, _, stderr := execInContainerOutput(t, ctx, c, cmd...)
	require.Equalf(t, 0, code, "command %v failed\nstderr: %s", cmd, stderr)
}

// execInContainerOutput runs a command in the container and returns the exit
// code, stdout and stderr.
func execInContainerOutput(t *testing.T, ctx context.Context, c testcontainers.Container, cmd ...string) (int, string, string) {
	t.Helper()

	code, reader, err := c.Exec(ctx, cmd, tcexec.WithWorkingDir("/app"))
	require.NoErrorf(t, err, "exec %v", cmd)

	// testcontainers multiplexes stdout+stderr into the returned reader; for our
	// assertions the combined stream is sufficient.
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	combined := buf.String()
	return code, combined, combined
}

// terminate stops and removes a container, logging (not failing) on error.
func terminate(t *testing.T, c testcontainers.Container) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := c.Terminate(ctx); err != nil {
		t.Logf("terminate container: %v", err)
	}
}

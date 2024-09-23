//nolint:funlen,lll
package integration_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration(t *testing.T) {
	require := require.New(t)

	// Using this to do some request to caddy.
	reqRoot := func() {
		r, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"http://127.0.0.1:2113",
			nil,
		)
		require.NoError(err)

		r.Header.Set("User-Agent", "integrationtest")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(err)

		defer resp.Body.Close()

		<-time.After(time.Second)
		t.Log("done request")
	}

	reqMetrics := func() string {
		r, err := http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			"http://127.0.0.1:2112/metrics",
			nil,
		)
		require.NoError(err)

		resp, err := http.DefaultClient.Do(r)
		require.NoError(err)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(err)

		return string(b)
	}

	// Volume for sharing logs.
	vol := testcontainers.ContainerMounts{
		{
			Source: testcontainers.DockerVolumeMountSource{
				Name: "logs",
			},
			Target: "/var/log/caddy",
		},
	}

	_, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			Started: true,
			ContainerRequest: testcontainers.ContainerRequest{
				Mounts:       vol,
				Image:        "caddy:2.8.4",
				ExposedPorts: []string{"2113:2113"},
				Files: []testcontainers.ContainerFile{
					{
						HostFilePath:      "Caddyfile",
						ContainerFilePath: "/etc/caddy/Caddyfile",
					},
				},
				WaitingFor: wait.ForHTTP("http://127.0.0.1:2113"),
			},
		},
	)
	require.NoError(err)

	reqRoot()
	<-time.After(time.Minute)

	_, err = testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			Started: true,
			ContainerRequest: testcontainers.ContainerRequest{
				Env: map[string]string{
					"CADDY_LOG_EXPORTER_LOG_FILES": "/var/log/caddy/caddy.log",
				},
				Mounts:       vol,
				Image:        "caddy-log-exporter:NOTUSE",
				ExposedPorts: []string{"2112:2112"},
				WaitingFor: wait.ForHTTP("http://127.0.0.1:2112/healthz").
					WithStartupTimeout(5 * time.Minute),
			},
		},
	)
	require.NoError(err)

	// Doing 5 requests.
	reqRoot()
	reqRoot()
	reqRoot()
	reqRoot()
	reqRoot()

	require.Contains(
		reqMetrics(),
		`caddy_log_exporter_http_request_total{host="127.0.0.1:2113", proto="HTTP/1.1", method="GET", status="200", user_agent="integrationtest"} 5`,
	)
}

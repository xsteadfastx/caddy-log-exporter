//nolint:gochecknoinits,lll,funlen
package server_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.xsfx.dev/caddy-log-exporter/internal/config"
	"go.xsfx.dev/caddy-log-exporter/internal/logger"
	"go.xsfx.dev/caddy-log-exporter/internal/server"
)

func init() {
	logger.Setup()
}

func TestServe(t *testing.T) {
	entry := `{"level":"info","ts":1726651920.3685553,"logger":"http.log.access.log0","msg":"handled request","request":{"remote_ip":"x.x.x.x","remote_port":"50440","client_ip":"x.x.x.x","proto":"HTTP/2.0","method":"GET","host":"foo.tld","uri":"/prometheus/cadvisor/src/commit/09f63dbfd2b38e76c204c8809fb4c2d1c9ad5879/vendor/github.com/golang/snappy/decode.go","headers":{"User-Agent":["facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"],"Accept":["*/*"]},"tls":{"resumed":false,"version":772,"cipher_suite":4865,"proto":"h2","server_name":"git.xsfx.dev"}},"bytes_read":0,"user_id":"","duration":0.003550188,"size":0,"status":200,"resp_headers":{"Server":["Caddy"],"Alt-Svc":["h3=\":443\"; ma=2592000"],"Content-Length":[""]}}`

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	require := require.New(t)

	td := t.TempDir()
	tf := path.Join(td, "foo.txt")

	file, err := os.Create(tf)
	require.NoError(err)
	t.Cleanup(func() { file.Close() })

	// Writing 10 log entries in a separated goroutine.
	go func(f *os.File) {
		<-time.After(time.Second) // Wait before we start writing log entries

		var counter int

		for range 10 {
			_, err := f.WriteString(entry)
			if err != nil {
				t.Logf("writing string: %s", err)

				return
			}

			counter++

			<-time.After(time.Second / 5)
		}
	}(file)

	// Creating server.
	s := server.New(config.Config{Addr: ":2112", LogFiles: []string{tf}})

	// Starting server serve.
	errChan := make(chan error)

	go func(ctx context.Context) {
		t.Log("serving")

		if err := s.Serve(ctx); err != nil {
			t.Logf("got error: %s", err)

			errChan <- err
		}
	}(ctx)

	// Wait if we get an error from serving.
	select {
	case err := <-errChan:
		t.Fatal(err)
	case <-time.After(time.Second):
	}

	t.Log("waiting for metrics")
	<-time.After(2 * time.Second)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"http://127.0.0.1:2112/metrics",
		nil,
	)
	require.NoError(err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(err)

	t.Cleanup(func() {
		require.NoError(resp.Body.Close())
	})

	b, err := io.ReadAll(resp.Body)
	require.NoError(err)

	require.Equal(
		`caddy_log_exporter_http_request_duration_seconds_bucket{host="foo.tld", proto="HTTP/2.0", method="GET", status="200", user_agent="facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",vmrange="3.162e-03...3.594e-03"} 10
caddy_log_exporter_http_request_duration_seconds_sum{host="foo.tld", proto="HTTP/2.0", method="GET", status="200", user_agent="facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"} 0.03550188
caddy_log_exporter_http_request_duration_seconds_count{host="foo.tld", proto="HTTP/2.0", method="GET", status="200", user_agent="facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"} 10
caddy_log_exporter_http_request_total{host="foo.tld", proto="HTTP/2.0", method="GET", status="200", user_agent="facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"} 10
`,
		string(b),
	)
}

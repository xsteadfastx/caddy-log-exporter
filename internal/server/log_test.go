//nolint:lll
package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.xsfx.dev/caddy-log-exporter/internal/server"
)

func TestParseLog(t *testing.T) {
	tables := []struct {
		name     string
		input    []byte
		expected server.Log
		err      error
	}{
		{
			name: "00",
			input: []byte(
				`{"level":"info","ts":1726651920.3685553,"logger":"http.log.access.log0","msg":"handled request","request":{"remote_ip":"x.x.x.x","remote_port":"50440","client_ip":"x.x.x.x","proto":"HTTP/2.0","method":"GET","host":"foo.tld","uri":"/prometheus/cadvisor/src/commit/09f63dbfd2b38e76c204c8809fb4c2d1c9ad5879/vendor/github.com/golang/snappy/decode.go","headers":{"User-Agent":["facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"],"Accept":["*/*"]},"tls":{"resumed":false,"version":772,"cipher_suite":4865,"proto":"h2","server_name":"git.xsfx.dev"}},"bytes_read":0,"user_id":"","duration":0.003550188,"size":0,"status":200,"resp_headers":{"Server":["Caddy"],"Alt-Svc":["h3=\":443\"; ma=2592000"],"Content-Length":[""]}}`,
			),
			expected: server.Log{
				Request: server.Request{
					RemoteIP: "x.x.x.x",
					Proto:    "HTTP/2.0",
					Method:   "GET",
					Host:     "foo.tld",
					URI:      "/prometheus/cadvisor/src/commit/09f63dbfd2b38e76c204c8809fb4c2d1c9ad5879/vendor/github.com/golang/snappy/decode.go",
					Headers: server.Headers{
						UserAgent: []string{
							"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
						},
					},
				},
				Duration: 0.003550188,
				Status:   200,
				Size:     0,
			},
			err: nil,
		},
	}

	require := require.New(t)

	for _, tt := range tables {
		t.Run(tt.name, func(_ *testing.T) {
			l, err := server.ParseLog(tt.input)
			if tt.err == nil {
				require.NoError(err)
				require.Equal(tt.expected, l)
			} else {
				require.ErrorIs(err, tt.err)
			}
		})
	}
}

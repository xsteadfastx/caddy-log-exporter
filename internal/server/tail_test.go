//nolint:lll,funlen
package server_test

import (
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/nxadm/tail"
	"github.com/stretchr/testify/require"
	"go.xsfx.dev/caddy-log-exporter/internal/config"
	"go.xsfx.dev/caddy-log-exporter/internal/server"
)

func TestTail(t *testing.T) {
	entry := `{"level":"info","ts":1726651920.3685553,"logger":"http.log.access.log0","msg":"handled request","request":{"remote_ip":"x.x.x.x","remote_port":"50440","client_ip":"x.x.x.x","proto":"HTTP/2.0","method":"GET","host":"foo.tld","uri":"/prometheus/cadvisor/src/commit/09f63dbfd2b38e76c204c8809fb4c2d1c9ad5879/vendor/github.com/golang/snappy/decode.go","headers":{"User-Agent":["facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)"],"Accept":["*/*"]},"tls":{"resumed":false,"version":772,"cipher_suite":4865,"proto":"h2","server_name":"git.xsfx.dev"}},"bytes_read":0,"user_id":"","duration":0.003550188,"size":0,"status":200,"resp_headers":{"Server":["Caddy"],"Alt-Svc":["h3=\":443\"; ma=2592000"],"Content-Length":[""]}}`
	require := require.New(t)

	td := t.TempDir()
	tf := path.Join(td, "foo.txt")

	file, err := os.Create(tf)
	require.NoError(err)
	t.Cleanup(func() { file.Close() })

	// Add one first line that should be ignored.
	_, err = file.WriteString(entry)
	require.NoError(err)

	wg := &sync.WaitGroup{}

	// Writing 10 log entries in a separated goroutine.
	wg.Add(1)

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

			<-time.After(time.Second)
		}

		wg.Done()
	}(file)

	s := server.New(config.Config{LogFiles: []string{tf}})

	var counter int

	m := sync.Mutex{}

	require.NoError(
		s.Tail(func(t *tail.Tail) {
			for range t.Lines {
				m.Lock()
				counter++
				m.Unlock()
			}
		}),
	)

	wg.Wait()

	m.Lock()
	require.Equalf(10, counter, "should counted to 10 even if we have 11 lines in the file")
	m.Unlock()
}

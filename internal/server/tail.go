package server

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/VictoriaMetrics/metrics"
	"github.com/nxadm/tail"
)

func (s *Server) Tail(f func(t *tail.Tail)) error {
	for _, l := range s.logFiles {
		t, err := tail.TailFile(l, tail.Config{
			ReOpen:    true,
			Follow:    true,
			MustExist: true,
			Location: &tail.SeekInfo{
				Offset: 0,
				Whence: io.SeekEnd, // Needed that it just uses new values
			},
		})
		if err != nil {
			return fmt.Errorf("tailing file: %w", err)
		}

		go f(t)
	}

	return nil
}

func parseFunc(t *tail.Tail) {
	for line := range t.Lines {
		l, err := ParseLog([]byte(line.Text))
		if err != nil {
			slog.Error("parsing log entry", "err", err)

			continue
		}

		metrics.GetOrCreateCounter(
			fmt.Sprintf(
				`%s_http_request_total{host="%s", proto="%s", method="%s", status="%d", user_agent="%s"}`,
				metricBaseName,
				l.Request.Host,
				l.Request.Proto,
				l.Request.Method,
				l.Status,
				l.Request.Headers.UserAgent[0],
			),
		).Inc()

		metrics.GetOrCreateHistogram(
			fmt.Sprintf(
				`%s_http_request_duration_seconds{host="%s", proto="%s", method="%s", status="%d", user_agent="%s"}`,
				metricBaseName,
				l.Request.Host,
				l.Request.Proto,
				l.Request.Method,
				l.Status,
				l.Request.Headers.UserAgent[0],
			),
		).Update(l.Duration)
	}
}

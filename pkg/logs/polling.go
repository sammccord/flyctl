package logs

import (
	"context"
	"time"

	"github.com/jpillora/backoff"
	"github.com/pkg/errors"
	"github.com/sammccord/flyctl/api"
)

type pollingStream struct {
	err       error
	apiClient *api.Client
}

func NewPollingStream(ctx context.Context, client *api.Client, opts *LogOptions) (LogStream, error) {
	_, err := client.GetApp(ctx, opts.AppName)
	if err != nil {
		return nil, errors.Wrap(err, "err polling logs")
	}
	return &pollingStream{apiClient: client}, nil
}

func (s *pollingStream) Stream(ctx context.Context, opts *LogOptions) <-chan LogEntry {
	out := make(chan LogEntry)

	b := &backoff.Backoff{
		Min:    250 * time.Millisecond,
		Max:    5 * time.Second,
		Factor: 2,
		Jitter: true,
	}

	if opts.MaxBackoff != 0 {
		b.Max = opts.MaxBackoff
	}

	go func() {
		defer close(out)
		errorCount := 0
		nextToken := ""

		var wait <-chan time.Time

		for {
			entries, token, err := s.apiClient.GetAppLogs(opts.AppName, nextToken, opts.RegionCode, opts.VMID)

			if err != nil {
				errorCount++

				if api.IsNotAuthenticatedError(err) || api.IsNotFoundError(err) || errorCount > 10 {
					s.err = err
					return
				}
				wait = time.After(b.Duration())
			} else {
				errorCount = 0

				if len(entries) == 0 {
					wait = time.After(b.Duration())
				} else {
					b.Reset()

					for _, entry := range entries {
						out <- LogEntry{
							Instance:  entry.Instance,
							Level:     entry.Level,
							Message:   entry.Message,
							Region:    entry.Region,
							Timestamp: entry.Timestamp,
							Meta:      entry.Meta,
						}
					}
					wait = time.After(0)

					if token != "" {
						nextToken = token
					}
				}
			}

			select {
			case <-ctx.Done():
				return
			case <-wait:
			}
		}
	}()

	return out
}

func (s *pollingStream) Err() error {
	return s.err
}

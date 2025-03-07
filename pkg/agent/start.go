//go:build !windows
// +build !windows

package agent

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sammccord/flyctl/api"
	"github.com/sammccord/flyctl/terminal"
)

func StartDaemon(ctx context.Context, api *api.Client, command string) (*Client, error) {
	startCh := make(chan error, 1)
	watchCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	cmd := exec.Command(command, "agent", "daemon-start")
	cmd.Env = append(os.Environ(), "FLY_NO_UPDATE_CHECK")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	agentPid := cmd.Process.Pid
	terminal.Debug("started agent process ", agentPid)

	// read stdout and stderr from the daemon process. If it
	// includes "[pid] OK" we know it started successfully, and
	// [pid] QUIT means it stopped. When it stops include the output with the
	// returnred error so it can be displayed to the user
	f, err := getLogFile(agentPid)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// tail the agent log until we see a status message or a timeout
	go func() {
		var output bytes.Buffer

		pidPrefix := fmt.Sprintf("[%d] ", agentPid)
		okPattern := pidPrefix + "OK"
		quitPattern := pidPrefix + "QUIT"

		var ok bool

	READ:
		for line := range tailReader(watchCtx, f) {
			switch {
			case strings.Contains(line, okPattern):
				ok = true
				break READ
			case strings.Contains(line, quitPattern):
				break READ
			default:
				if strings.Contains(line, pidPrefix) {
					if output.Len() > 0 {
						output.WriteByte(byte('\n'))
					}
					output.WriteString(line)
				}
			}
		}

		if ok {
			startCh <- nil
			return
		}

		startCh <- &AgentStartError{Output: output.String()}
	}()

	// wait for the output to include a running or failed message
	if startErr := <-startCh; startErr != nil {
		return nil, startErr
	}

	client, err := waitForClient(ctx, api)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't establish connection to Fly Agent")
	}

	return client, nil
}

func waitForClient(ctx context.Context, api *api.Client) (*Client, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	respCh := make(chan *Client, 1)

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)

			c, err := DefaultClient(api)
			if err == nil {
				_, err := c.Ping(ctx)
				if err == nil {
					respCh <- c
					break
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case client := <-respCh:
		return client, nil
	}
}

// naive tail implementation
func tailReader(ctx context.Context, r io.Reader) <-chan string {
	out := make(chan string)

	pr, pw := io.Pipe()

	go func() {
		defer close(out)

		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()

	go func() {
		defer pw.Close()

		buf := make([]byte, 1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				pw.Write(buf[:n])
			}
			if errors.Is(err, io.EOF) {
				time.Sleep(100 * time.Millisecond)
			}
			if ctx.Err() != nil {
				break
			}
		}
	}()

	return out
}

package dispatch

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/saga-changelog/saga/internal/theme"
	"github.com/saga-changelog/saga/pkg/courier"
)

// BinaryName returns the expected binary name for a courier.
func BinaryName(name string) string {
	return "saga-courier-" + name
}

// LookupBinary checks if a courier binary is available on PATH.
func LookupBinary(name string) (string, error) {
	return exec.LookPath(BinaryName(name))
}

// envForCourier returns a filtered copy of the current environment.
// It includes all env vars except SAGA_COURIER_* vars that don't match
// the given courier name. This prevents leaking secrets between couriers.
//
// The name/key boundary uses a double underscore, so a courier named
// "stdout" sees only SAGA_COURIER_STDOUT__* vars. This disambiguates
// couriers whose canonical name is a prefix of another (e.g. "google"
// vs "google-chat").
func envForCourier(name string) []string {
	prefix := "SAGA_COURIER_" + strings.ToUpper(strings.ReplaceAll(name, "-", "_")) + "__"
	const globalPrefix = "SAGA_COURIER_"

	var filtered []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, globalPrefix) {
			if !strings.HasPrefix(env, prefix) {
				continue
			}
		}
		filtered = append(filtered, env)
	}
	return filtered
}

// Info invokes the courier's "info" subcommand and returns the parsed result.
func Info(name string) (*courier.Info, error) {
	bin, err := LookupBinary(name)
	if err != nil {
		return nil, fmt.Errorf("courier %q not found on PATH", name)
	}

	cmd := exec.Command(bin, "info")
	cmd.Env = envForCourier(name)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("courier %q info failed: %s", name, stderr.String())
	}

	var info courier.Info
	if err := json.Unmarshal(stdout.Bytes(), &info); err != nil {
		return nil, fmt.Errorf("courier %q info returned invalid JSON: %w", name, err)
	}

	return &info, nil
}

// ValidateRoute invokes the courier's "validate-route" subcommand with the
// given route. This checks credentials and connectivity without any tale content.
func ValidateRoute(name string, route *courier.Route) error {
	bin, err := LookupBinary(name)
	if err != nil {
		return fmt.Errorf("courier %q not found on PATH", name)
	}

	routeJSON, err := json.Marshal(route)
	if err != nil {
		return fmt.Errorf("marshaling route: %w", err)
	}

	cmd := exec.Command(bin, "validate-route")
	cmd.Env = envForCourier(name)
	cmd.Stdin = bytes.NewReader(routeJSON)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("courier %q validate-route failed: %s", name, stderr.String())
	}

	return nil
}

// Tell invokes the courier's "tell" subcommand with the given payload.
// Stdout and stderr from the courier are prefixed with the courier name.
func Tell(name string, p *courier.Payload) error {
	bin, err := LookupBinary(name)
	if err != nil {
		return fmt.Errorf("courier %q not found on PATH", name)
	}

	payloadJSON, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	cmd := exec.Command(bin, "tell")
	cmd.Env = envForCourier(name)
	cmd.Stdin = bytes.NewReader(payloadJSON)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("creating stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting courier %q: %w", name, err)
	}

	styledName := theme.CourierName.Render(name)
	styledStdoutSep := theme.StdoutSep.Render(" | ")
	styledStderrSep := theme.StderrSep.Render(" ! ")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		prefixLines(stdoutPipe, os.Stdout, styledName, styledStdoutSep)
	}()
	var stderrBuf bytes.Buffer
	go func() {
		defer wg.Done()
		tee := io.TeeReader(stderrPipe, &stderrBuf)
		prefixLines(tee, os.Stderr, styledName, styledStderrSep)
	}()
	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("courier %q tell failed: %s", name, stderrBuf.String())
	}

	return nil
}

// prefixLines reads lines from r and writes them to w with a styled prefix.
func prefixLines(r io.Reader, w io.Writer, styledName, styledSep string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Fprintf(w, "%s%s %s\n", styledName, styledSep, scanner.Text())
	}
}

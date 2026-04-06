package courier

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Run parses os.Args, dispatches to the appropriate Courier method, and
// exits. It never returns. Intended to be called from main().
//
// On SIGINT or SIGTERM, the context passed to ValidateRoute and Tell is
// cancelled so handlers can abort cleanly.
func Run(c Courier) {
	if c == nil {
		fmt.Fprintln(os.Stderr, "courier.Run: nil Courier")
		os.Exit(2)
	}

	info := c.Info()
	if info.Name == "" {
		fmt.Fprintln(os.Stderr, "courier.Run: Courier.Info().Name is required")
		os.Exit(2)
	}

	if len(os.Args) < 2 {
		usage(info.Name)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch os.Args[1] {
	case "info":
		if err := json.NewEncoder(os.Stdout).Encode(info); err != nil {
			fmt.Fprintf(os.Stderr, "encoding info: %v\n", err)
			os.Exit(1)
		}

	case "validate-route":
		route, err := decodeRoute()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := c.ValidateRoute(ctx, route); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	case "tell":
		p, err := decodePayload()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := c.Tell(ctx, p); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

	default:
		usage(info.Name)
		os.Exit(1)
	}
}

func decodeRoute() (*Route, error) {
	var r Route
	if err := json.NewDecoder(os.Stdin).Decode(&r); err != nil {
		return nil, fmt.Errorf("decoding route: %w", err)
	}
	return &r, nil
}

func decodePayload() (*Payload, error) {
	var p Payload
	if err := json.NewDecoder(os.Stdin).Decode(&p); err != nil {
		return nil, fmt.Errorf("decoding payload: %w", err)
	}
	return &p, nil
}

func usage(name string) {
	fmt.Fprintf(os.Stderr, "Usage: saga-courier-%s {info|validate-route|tell}\n", name)
}

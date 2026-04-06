package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/saga-changelog/saga/internal/safejsonnet"
)

// Config is the top-level Saga configuration.
type Config struct {
	Audiences []Audience `json:"audiences"`
}

// Audience defines a target group that receives tales.
type Audience struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Interest    string  `json:"interest"`
	Tone        string  `json:"tone"`
	Routes      []Route `json:"routes"`
}

// Route defines a delivery path to a courier.
type Route struct {
	Name    string       `json:"name"`
	Courier RouteCourier `json:"courier"`
}

// RouteCourier identifies the courier binary and its config for a route.
type RouteCourier struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

// AudienceRoute pairs a route with its owning audience name.
type AudienceRoute struct {
	Audience string
	Route    Route
}

// QualifiedName returns the route identifier as "audience/route".
func (ar AudienceRoute) QualifiedName() string {
	return ar.Audience + "/" + ar.Route.Name
}

// AudienceByName returns the audience with the given name, or false if not found.
func (c *Config) AudienceByName(name string) (Audience, bool) {
	for _, a := range c.Audiences {
		if a.Name == name {
			return a, true
		}
	}
	return Audience{}, false
}

// AllRoutes returns all routes across all audiences, each paired with its audience name.
func (c *Config) AllRoutes() []AudienceRoute {
	var all []AudienceRoute
	for _, a := range c.Audiences {
		for _, r := range a.Routes {
			all = append(all, AudienceRoute{Audience: a.Name, Route: r})
		}
	}
	return all
}

// Load reads and evaluates the config.jsonnet file from the given saga directory.
func Load(sagaDir string) (*Config, error) {
	configPath := filepath.Join(sagaDir, "config.jsonnet")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	vm, err := safejsonnet.MakeVM(sagaDir)
	if err != nil {
		return nil, fmt.Errorf("creating jsonnet VM: %w", err)
	}
	jsonStr, err := vm.EvaluateAnonymousSnippet(configPath, string(data))
	if err != nil {
		return nil, fmt.Errorf("evaluating config jsonnet: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal([]byte(jsonStr), &cfg); err != nil {
		return nil, fmt.Errorf("deserializing config: %w", err)
	}

	return &cfg, nil
}

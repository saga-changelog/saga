package courier

import (
	"context"
)

// Courier is the interface that courier plugin implementations satisfy.
// Each method corresponds to a subcommand of the courier binary.
type Courier interface {
	// Info returns static metadata about the courier. Called for the
	// "info" subcommand. Should be cheap and side-effect free.
	Info() Info

	// ValidateRoute verifies the courier can deliver to the given route
	// (credentials present, endpoints reachable, config valid, etc.)
	// without actually delivering anything. Called for the "validate-route"
	// subcommand. Receives only the route, not any tale content.
	ValidateRoute(ctx context.Context, route *Route) error

	// Tell delivers the tale in the given payload to the courier's platform.
	// Called for the "tell" subcommand.
	Tell(ctx context.Context, p *Payload) error
}

// Info is the static description of a courier returned from Courier.Info
// and emitted by the "info" subcommand.
type Info struct {
	// Name is the courier's identifier (e.g., "slack"). The binary should
	// be named saga-courier-<Name>.
	Name string `json:"name"`

	// Description is a human-readable description of the courier.
	Description string `json:"description"`

	// ConfigKeys describes the configuration keys this courier accepts
	// in a route's config map, including validation regex.
	ConfigKeys []ConfigKey `json:"config_keys"`
}

// ConfigKey describes a single configuration key that a courier accepts
// in a route's config map.
type ConfigKey struct {
	// Name is the config key name (e.g., "channel", "project_id").
	Name string `json:"name"`

	// Description is a human-readable description of the key.
	Description string `json:"description"`

	// Regex is a regular expression that the config value must match.
	// The regex determines whether the key is required: if the regex
	// matches an empty string, the key is optional; if it does not,
	// the key is required. An empty regex means any value is accepted.
	Regex string `json:"regex,omitempty"`
}

// Payload is the structure passed to a courier binary on stdin for the
// "tell" subcommand. Each invocation carries a single tale: saga invokes
// the courier once per tale to deliver, so each call produces one message
// on the destination platform.
type Payload struct {
	// Chapter is the version string of the chapter being told.
	Chapter string `json:"chapter"`

	// Route contains the route that triggered this courier invocation.
	Route Route `json:"route"`

	// Tale is the single tale to deliver on this invocation.
	Tale Tale `json:"tale"`
}

// Route identifies which route triggered the delivery and provides the
// courier-specific configuration for it.
type Route struct {
	// Name is the route's name (unique within its audience).
	Name string `json:"name"`

	// Audience is the audience this route serves.
	Audience string `json:"audience"`

	// Config is the courier-specific configuration from the route
	// (e.g., channel name for Slack, project ID for Basecamp).
	Config map[string]string `json:"config"`
}

// Tale is a single audience-specific description of a feat to deliver.
type Tale struct {
	// Feat is the slug/filename of the feat (without extension).
	Feat string `json:"feat"`

	// Title is a plain-text, single-line title for this tale. It is
	// used as the subject/header of the resulting message.
	Title string `json:"title"`

	// Text is the structured text of the tale. Saga parses the
	// author's markdown once and emits this neutral representation so
	// couriers can render it into their platform's native format
	// without pulling in a markdown parser.
	Text TaleText `json:"text"`
}

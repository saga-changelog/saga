# Saga — Plan Document

This document describes the over-arching plan for the new `saga` tool.

## General approach

While we work/iterate through this document, compact it as we go. Once a technical implementation detail has been converted to actual code in this repo, we can remove the instructions to do so from this file.

## Summary

Saga is a CLI tool for managing and communicating software changes across internal and external audiences. It replaces the traditional `CHANGELOG.md` approach with a structured, file-per-change system that not only tracks what changed, but actively delivers tailored notifications to the right people through the right channels after deployment.

The core innovation is the separation of concerns: developers describe what changed and for whom, operators configure where those descriptions should be delivered, and the tool handles the routing and dispatch. This means a single change — like a billing invoice restructuring — can be communicated in technical language to the internal finance team on Slack, and in customer-friendly language on an external blog, all from one file written by the developer.

### Core Concepts

| Concept | Term | Description |
|---------|------|-------------|
| The tool | **Saga** | CLI tool, single Go binary |
| A change record | **Feat** | A Jsonnet file describing one accomplishment, containing one or more tales for different audiences |
| Audience-specific text | **Tale** | The description of a change written for a specific audience |
| Pending feats | **Pending** | Feats not yet assigned to a chapter |
| A release bundle | **Chapter** | A versioned collection of feats |
| Target groups | **Audiences** | Named groups that receive tales (e.g., billing, partners, engineering) |
| Routing configs | **Routes** | Rules mapping audiences to couriers with delivery config |
| Delivery plugins | **Couriers** | External binaries that deliver tales to specific platforms |
| Publishing | **Tell** | The act of dispatching couriers after deployment |

### Technology Stack

| Layer | Technology | Notes |
|-------|-----------|-------|
| Config format | Jsonnet | Single config language for both config and feats |
| Feat format | Jsonnet | Same language, consistent developer experience |
| Core tool | Go | Compiles to a single static binary |
| CLI framework | alecthomas/kong | Struct-based CLI parser |
| Jsonnet evaluation | google/go-jsonnet | Pure Go, no CGO, no external binary needed |
| Validation | Go code | Hand-written structs and validation logic |
| Couriers | External binaries | Language-agnostic plugin system |
| Version control | Git | Hard dependency |

### Hard Dependencies

- **Git**: Saga requires a git repository. The directory structure, feat management, and chapter creation are all designed around git workflows.
- **Go**: The tool is written in Go and distributed as a single static binary.

### Optional Dependencies

- **Couriers**: Each courier is an optional external binary. Saga functions fully without any couriers installed — you can still create feats, validate, and create chapters. Couriers are only needed for the `saga tell` step.

---

## Directory Structure

### Saga Project Directory (inside a user's repository)

The Saga data directory is `.saga/` at the root of a git repository. The dot prefix follows the convention of other development tools (`.github/`, `.gitlab/`, `.husky/`, etc.).

```
.saga/
  config.jsonnet                          # Main configuration: audiences and routes
  feats/
    pending/                              # Feats waiting to be included in a chapter
      invoice-restructuring.jsonnet
      dashboard-redesign.jsonnet
    chapters/                             # Released chapters, named by version
      2.3.0/
        api-rate-limits.jsonnet
        new-auth-flow.jsonnet
      2.4.0/
        invoice-restructuring.jsonnet
        dashboard-redesign.jsonnet
```

Saga has no opinion on chapter name structure. Although it is strongly advised to use a scheme that can be easily sorted.

### Conventions

- Feat filenames use kebab-case slugs: `invoice-restructuring.jsonnet`
- Chapter directories are named after the version string exactly as provided by the user
- Saga does not enforce semver — any string is accepted as a version
- The `.saga/` directory must live at the root of a git repository

### Saga Source Repository Structure

The Saga tool itself is a Go project. `main.go` lives at the repository root so that the tool is easily installable with `go install`.

```
saga/                                     # The Saga tool repository
  main.go                                 # Entry point — at root for go install
  internal/
    config/
      config.go                           # Config loading and Go types
      validation.go                       # Config validation logic
    feat/
      feat.go                             # Feat loading and Go types
      validation.go                       # Feat validation logic
    chapter/
      chapter.go                          # Chapter creation and management
    courier/
      discovery.go                        # Courier binary discovery on PATH
      runner.go                           # Courier invocation and result handling
    git/
      git.go                              # Git operations (repo detection, etc.)
    cli/
      init.go                             # saga init
      validate.go                         # saga validate
      pending.go                          # saga pending
      chapters.go                         # saga chapters / saga chapters create
      tell.go                             # saga tell
      audiences.go                        # saga audiences
      routes.go                           # saga routes
      couriers.go                         # saga couriers
  pkg/
    courier/
      payload.go                          # Courier payload types — the contract between Saga and couriers
  go.mod
  go.sum
```

---

## Courier Payload Package (`pkg/courier`)

This is a public Go package that defines the intermediate data format between the Saga main tool and courier plugins. It lives in `pkg/` (not `internal/`) so that courier authors writing their plugins in Go can import it directly.

This package contains the types that are serialized to JSON and passed to courier binaries on stdin. By publishing these as a Go package, courier authors get type safety and don't need to reverse-engineer the JSON contract.

### Types

```go
package courier

// Payload is the top-level structure passed to a courier binary on stdin.
type Payload struct {
    // Chapter is the version string of the chapter being told.
    Chapter string `json:"chapter"`

    // Route contains the route that triggered this courier invocation.
    Route PayloadRoute `json:"route"`

    // Tales is the list of tales to deliver, one per feat that has
    // a tale for this route's audience.
    Tales []PayloadTale `json:"tales"`
}

// PayloadRoute identifies which route triggered the delivery
// and provides the courier-specific configuration.
type PayloadRoute struct {
    // Name is the unique name of the route.
    Name string `json:"name"`

    // Audience is the audience this route serves.
    Audience string `json:"audience"`

    // Config is the courier-specific configuration from the route
    // (e.g., channel name for Slack, site URL for WordPress).
    Config map[string]string `json:"config"`
}

// PayloadTale is a single tale to deliver, representing one feat's
// audience-specific text.
type PayloadTale struct {
    // Feat is the slug/filename of the feat (without extension).
    Feat string `json:"feat"`

    // Summary is the one-line summary from the feat.
    Summary string `json:"summary"`

    // Text is the audience-specific tale text.
    Text string `json:"text"`
}

// CourierInfo is returned by a courier binary when invoked with the "info" subcommand.
type CourierInfo struct {
    // Name is the courier's identifier.
    Name string `json:"name"`

    // Description is a human-readable description of the courier.
    Description string `json:"description"`

    // ConfigKeys describes the configuration keys the courier accepts,
    // including validation regex and whether each key is required.
    ConfigKeys []CourierConfigKey `json:"config_keys"`
}

// CourierConfigKey describes a single configuration key for a courier.
type CourierConfigKey struct {
    // Name is the config key name (e.g., "channel", "site").
    Name string `json:"name"`

    // Description is a human-readable description of the key.
    Description string `json:"description"`

    // Regex is a regular expression that the config value must match.
    // The regex determines whether the key is required: if the regex
    // matches an empty string, the key is optional; if it does not,
    // the key is required. An empty regex means any value is accepted
    // (the key is purely informational).
    Regex string `json:"regex,omitempty"`
}
```

### Usage by Courier Authors

A Go-based courier can import this package:

```go
import "github.com/saga-changelog/saga/pkg/courier"

func main() {
    switch os.Args[1] {
    case "info":
        // ... output CourierInfo JSON
    case "tell":
        var payload courier.Payload
        json.NewDecoder(os.Stdin).Decode(&payload)
        // ... deliver the tales
    }
}
```

Courier authors using other languages simply follow the JSON schema that these types define.

---

## Configuration

The main configuration file is `.saga/config.jsonnet`. It defines two things: the audiences that can receive tales, and the routes that deliver tales to couriers.

### Configuration Schema

```jsonnet
// .saga/config.jsonnet
{
  audiences: [
    {
      name: "billing",
      description: "Internal finance department.",
      interest: "Billing cycle changes, invoice formats, tax handling, payment terms.",
      tone: "Direct and technical. Reference ticket numbers.",
    },
    {
      name: "partners",
      description: "External partners and clients.",
      interest: "Product improvements, pricing changes, new integrations.",
      tone: "Friendly and benefit-focused. No internal jargon.",
    },
    {
      name: "engineering",
      description: "Internal engineering team.",
      interest: "API changes, schema migrations, breaking changes, performance improvements.",
      tone: "Technical. Reference APIs, schemas, and endpoints.",
    },
  ],

  routes: [
    {
      name: "billing-slack",
      audience: "billing",
      courier: {
        name: "slack",
        config: {
          channel: "#finance-updates",
        },
      },
    },
    {
      name: "partner-blog",
      audience: "partners",
      courier: {
        name: "wordpress",
        config: {
          site: "blog.example.com",
          category: "changelog",
        },
      },
    },
    {
      name: "partner-email",
      audience: "partners",
      courier: {
        name: "email",
        config: {
          template: "partner-update",
          list: "partner-notifications",
        },
      },
    },
    {
      name: "engineering-slack",
      audience: "engineering",
      courier: {
        name: "slack",
        config: {
          channel: "#engineering",
        },
      },
    },
  ],
}
```

### Go Struct Definitions

```go
type Config struct {
    Audiences []Audience `json:"audiences"`
    Routes    []Route    `json:"routes"`
}

type Audience struct {
    Name        string `json:"name"`
    Description string `json:"description"`
    Interest    string `json:"interest"`
    Tone        string `json:"tone"`
}

type Route struct {
    Name     string              `json:"name"`
    Audience string              `json:"audience"`
    Courier  RouteCourier `json:"courier"`
}

type RouteCourier struct {
    Name   string            `json:"name"`
    Config map[string]string `json:"config"`
}
```

### Configuration Validation Rules

When `saga validate` evaluates the config, it must check:

1. Every route's `audience` field references a name that exists in `audiences`
2. Every route's `courier.name` field references an installed courier binary (error if not found)
3. Every route has a unique `name`
4. Every courier binary must respond to the `info` subcommand (error if it fails)
5. Config values in each route match their courier's regex patterns (as reported by `info`); keys whose regex does not match an empty string are required

---

## Feat Format

Feats are Jsonnet files containing one or more tales, each targeting a specific audience. A feat represents a single change or feature; the tales within it are how that change is communicated to different groups.

### Feat Schema

```jsonnet
// .saga/feats/pending/invoice-restructuring.jsonnet
{
  summary: "Invoices now include itemized tax breakdowns",

  tales: [
    {
      audience: "billing",
      text: |||
        The invoice restructuring from PROJ-1234 is live.
        Invoices now include per-region tax breakdowns.
        Finance should see the new format starting next
        billing cycle. Questions? Head to #billing.
      |||,
    },
    {
      audience: "partners",
      text: |||
        We've improved our invoicing! Your invoices now
        include clear, itemized tax breakdowns for each
        region, making your bookkeeping easier.
      |||,
    },
  ],
}
```

### Minimal Feat

A feat needs only a summary and at least one tale:

```jsonnet
{
  summary: "Fixed dashboard loading performance",

  tales: [
    {
      audience: "engineering",
      text: |||
        Dashboard initial load time reduced from 3.2s
        to 0.8s by lazy-loading chart components.
      |||,
    },
  ],
}
```

### Advanced Feat Using Jsonnet Features

Jsonnet's expressiveness allows developers to share content between tales when appropriate:

```jsonnet
{
  summary: "Invoices now include itemized tax breakdowns",

  local detail = "per-region itemized tax breakdowns",

  tales: [
    {
      audience: "billing",
      text: |||
        The invoice restructuring is live. Invoices now
        include %(detail)s. Finance should see the new
        format starting next billing cycle.
      ||| % { detail: detail },
    },
    {
      audience: "partners",
      text: |||
        We've improved our invoicing! Your invoices now
        include %(detail)s, making your bookkeeping easier.
      ||| % { detail: detail },
    },
  ],
}
```

### Go Struct Definition

```go
type Feat struct {
    Summary string `json:"summary"`
    Tales   []Tale `json:"tales"`
}

type Tale struct {
    Audience string `json:"audience"`
    Text     string `json:"text"`
}
```

### Feat Validation Rules

When `saga validate` evaluates feats, it must check:

1. Every feat has a non-empty `summary`
2. Every feat has at least one entry in `tales`
3. Every tale has a non-empty `audience` that references a defined audience from `config.jsonnet`
4. Every tale has non-empty `text`
5. No duplicate `audience` values within a single feat's `tales`

---

## CLI Interface

### Design Principles

- Struct-based command definitions using `alecthomas/kong`
- Bare resource commands default to listing (e.g., `saga chapters` lists chapters)
- Action subcommands are explicit (e.g., `saga chapters create <version>`)
- No file-creation commands for feats — developers create Jsonnet files directly
- Clear exit codes: 0 for success, non-zero for errors
- Human-readable output by default

### Kong CLI Struct

```go
package main

import "github.com/alecthomas/kong"

type CLI struct {
    Init          InitCmd          `cmd:"" help:"Initialize Saga in the current git repository."`
    Validate      ValidateCmd      `cmd:"" help:"Validate all configuration and feat files."`
    Pending       PendingCmd       `cmd:"" help:"List pending feats."`
    Chapters      ChaptersCmd      `cmd:"" help:"List chapters or manage them."`
    Tell          TellCmd          `cmd:"" help:"Dispatch couriers for a chapter."`
    Audiences     AudiencesCmd     `cmd:"" help:"List configured audiences."`
    Routes RoutesCmd `cmd:"" help:"List configured routes."`
    Couriers      CouriersCmd      `cmd:"" help:"List available courier plugins."`
}

type InitCmd struct{}

type ValidateCmd struct{}

type PendingCmd struct{}

type ChaptersCmd struct {
    Create ChaptersCreateCmd `cmd:"" help:"Bundle pending feats into a new chapter."`
}

type ChaptersCreateCmd struct {
    Version  string   `arg:"" help:"Version string for the new chapter."`
    Feats []string `arg:"" optional:"" help:"Specific feat names to include. If omitted, all pending feats are included."`
}

type TellCmd struct {
    Version  string   `arg:"" help:"Chapter version to tell."`
    DryRun   bool     `help:"Preview what would be sent without dispatching." name:"dry-run"`
    Audience []string `help:"Only tell tales for specific audiences (mutually exclusive with --route)." name:"audience"`
    Route    []string `help:"Only dispatch specific routes by name (mutually exclusive with --audience)." name:"route"`
}

type AudiencesCmd struct{}

type RoutesCmd struct{}

type CouriersCmd struct{}

func main() {
    cli := CLI{}
    ctx := kong.Parse(&cli,
        kong.Name("saga"),
        kong.Description("Communicate software changes to the right people through the right channels."),
        kong.UsageOnError(),
    )
    err := ctx.Run()
    ctx.FatalIfErrorf(err)
}
```

### Command Reference

#### `saga init`

Initialize Saga in the current git repository.

**Behavior:**
- Verify the current directory is within a git repository; fail with an error if not
- Fail if `.saga/` directory already exists
- Create the directory structure: `.saga/feats/pending/`, `.saga/feats/chapters/`
- Create a skeleton `.saga/config.jsonnet` with commented examples
- Print a success message with next steps

**Exit codes:**
- `0` — initialization successful
- `1` — not a git repository, or Saga already initialized

---

#### `saga validate`

Validate all configuration and feat files.

**Behavior:**
- Evaluate `.saga/config.jsonnet` through go-jsonnet and deserialize into Go structs
- Validate config: all routes reference defined audiences, all route names are unique
- Evaluate every `.jsonnet` file in `.saga/feats/pending/` and `.saga/feats/chapters/*/`
- Validate each feat: summary is non-empty, tales array is non-empty, all tale audiences reference defined audiences
- Print all errors found (do not stop at the first error)
- Error if routes reference couriers that are not installed or fail the `info` subcommand

**Exit codes:**
- `0` — all files valid
- `1` — one or more validation errors

---

#### `saga pending`

List all feats currently in `.saga/feats/pending/`.

**Behavior:**
- Evaluate each pending feat through go-jsonnet
- Display each feat's filename, summary, and which audiences have tales
- If no pending feats exist, print a message indicating that

**Example output:**

```
Pending feats:

  invoice-restructuring
    "Invoices now include itemized tax breakdowns"
    audiences: billing, partners

  dashboard-redesign
    "Dashboard redesign with new navigation"
    audiences: engineering, partners

2 feats pending
```

---

#### `saga chapters`

List all chapters. This is the default behavior when no subcommand is given.

**Behavior:**
- Scan `.saga/feats/chapters/` for version directories
- For each chapter, count the number of feats
- Sort by directory name (or by creation date from git history)

**Example output:**

```
Chapters:

  2.3.0  (2 feats)
  2.4.0  (3 feats)

2 chapters
```

---

#### `saga chapters create <version> [feats...]`

Bundle pending feats into a new chapter.

**Arguments:**
- `<version>` — the version string for this chapter (any non-empty string)
- `[feats...]` — optional list of specific feat filenames to include. If omitted, all pending feats are included.

**Behavior:**
- Fail if the chapter version directory already exists
- Fail if no pending feats exist (or if specified feats are not found)
- Run validation on all feats being moved
- Move feat files from `.saga/feats/pending/` to `.saga/feats/chapters/<version>/`
- Do NOT create a git commit — let the developer commit as part of their release workflow
- Print a summary of what was moved

**Exit codes:**
- `0` — chapter created successfully
- `1` — version already exists, no pending feats, or validation error

---

#### `saga tell <version>`

Dispatch couriers for a chapter's tales.

**Arguments:**
- `<version>` — the chapter version to tell

**Flags:**
- `--dry-run` — preview what would be sent without dispatching any couriers
- `--audience <name>` — only tell tales for specific audiences (repeatable, mutually exclusive with `--route`)
- `--route <name>` — only dispatch specific routes by name (repeatable, mutually exclusive with `--audience`)

**Behavior:**
1. Load config and all feats from `.saga/feats/chapters/<version>/`
2. Run validation on config and feats (fail early if invalid)
3. Fail if both `--audience` and `--route` are specified
4. For each audience that has tales in this chapter, find all matching routes
5. If `--audience` is specified, filter to only those audiences
6. If `--route` is specified, filter to only those routes
7. For each route, gather all tales for its audience across all feats
8. Invoke `validate-config` on all involved couriers (fail early if any courier rejects its config)
9. Invoke the courier `tell` subcommand for each route
10. Report results

**Dry-run output example:**

```
Chapter 2.4.0 — dry run

  Route: billing-slack
    Courier: slack → #finance-updates
    Tales:
      - invoice-restructuring: "The invoice restructuring from..."
      - payment-terms-update: "Payment terms have been..."

  Route: partner-blog
    Courier: wordpress → blog.example.com
    Tales:
      - invoice-restructuring: "We've improved our invoicing!..."

  Route: partner-email
    Courier: email → partner-notifications
    Tales:
      - invoice-restructuring: "We've improved our invoicing!..."

3 routes, 2 couriers, 4 deliveries
```

**Exit codes:**
- `0` — all couriers dispatched successfully
- `1` — validation failed, or one or more courier dispatches failed

---

#### `saga audiences`

List all configured audiences with their descriptions. This is the default behavior when no subcommand is given.

**Behavior:**
- Load and evaluate `.saga/config.jsonnet`
- Display each audience name and its description
- This is the reference developers check before writing feats

**Example output:**

```
Audiences:

  billing
    Internal finance department.
    interest:  Billing cycle changes, invoice formats, tax handling, payment terms.
    tone:      Direct and technical. Reference ticket numbers.

  partners
    External partners and clients.
    interest:  Product improvements, pricing changes, new integrations.
    tone:      Friendly and benefit-focused. No internal jargon.

  engineering
    Internal engineering team.
    interest:  API changes, schema migrations, breaking changes, performance improvements.
    tone:      Technical. Reference APIs, schemas, and endpoints.

3 audiences
```

---

#### `saga routes`

List all configured routes. This is the default behavior when no subcommand is given.

**Behavior:**
- Load and evaluate `.saga/config.jsonnet`
- Display each route with its audience, courier, and config

**Example output:**

```
Routes:

  billing-slack
    audience: billing
    courier: slack
    config: channel=#finance-updates

  partner-blog
    audience: partners
    courier: wordpress
    config: site=blog.example.com, category=changelog

  partner-email
    audience: partners
    courier: email
    config: template=partner-update, list=partner-notifications

3 routes
```

---

#### `saga couriers`

List available courier plugins. This is the default behavior when no subcommand is given.

**Behavior:**
- Scan PATH for binaries matching the naming convention `saga-courier-*`
- For each found courier, extract the courier name from the binary name
- Invoke the courier with the `info` subcommand to get its description and config keys (show an error if `info` fails)
- Show installed status relative to what routes reference

**Example output:**

```
Couriers:

  slack       saga-courier-slack       installed
  wordpress   saga-courier-wordpress   installed
  email       saga-courier-email       not found (required by: partner-email)

2 installed, 1 missing
```

---

## Courier Plugin Contract

Couriers are external binaries that deliver tales to specific platforms. They follow a simple contract that makes them easy to build in any language.

### Naming Convention

Courier binaries are named `saga-courier-<name>` and must be available on PATH. For example:
- `saga-courier-slack`
- `saga-courier-wordpress`
- `saga-courier-basecamp`
- `saga-courier-email`

This follows the same pattern Git uses for subcommands (`git-lfs`, `git-credential-manager`).

### Subcommands

Couriers must support three subcommands:

- **`info`** — output a JSON description of the courier (see Info Subcommand section)
- **`validate-config`** — accept a JSON payload on stdin and validate that the courier can operate (credentials, connectivity, etc.) without actually delivering anything
- **`tell`** — accept a JSON payload on stdin and deliver the tales

### Invocation

Saga invokes a courier's `tell` subcommand with the JSON payload on stdin:

```bash
echo '<json payload>' | saga-courier-slack tell
```

### Input Payload

The courier receives a JSON object on stdin when invoked with `tell`. The structure is defined by the `pkg/courier` package (see Courier Payload Package section). Example:

```json
{
  "chapter": "2.4.0",
  "route": {
    "name": "billing-slack",
    "audience": "billing",
    "config": {
      "channel": "#finance-updates"
    }
  },
  "tales": [
    {
      "feat": "invoice-restructuring",
      "summary": "Invoices now include itemized tax breakdowns",
      "text": "The invoice restructuring from PROJ-1234 is live..."
    },
    {
      "feat": "payment-terms-update",
      "summary": "Payment terms extended to net-60",
      "text": "Payment terms have been extended..."
    }
  ]
}
```

Key design decisions:
- A courier receives ALL tales for its route's audience in a single invocation, not one invocation per tale. This lets the courier decide how to aggregate — a Slack courier might post one message with sections, a blog courier might create one post with bullet points.
- The `config` map contains courier-specific configuration from the route. For Slack, that's a channel. For WordPress, it might be a site URL and category.
- The `text` field contains the audience-specific tale text from the feat.

### Output Contract (tell)

- **Exit code 0**: delivery successful
- **Exit code non-zero**: delivery failed
- **Stderr**: error messages (Saga captures and displays these to the user)
- **Stdout**: optional, can be used for informational messages (Saga may display these to the user)

### Validate-Config Subcommand

Before dispatching any `tell` calls, Saga invokes `validate-config` on each involved courier. This lets the courier verify that credentials are available, endpoints are reachable, and the provided config is operationally valid — beyond what regex can check.

The courier receives the same JSON payload as `tell` on stdin:

```bash
echo '<json payload>' | saga-courier-slack validate-config
```

### Output Contract (validate-config)

- **Exit code 0**: config is valid, courier is ready to deliver
- **Exit code non-zero**: config is invalid or courier cannot operate (missing credentials, unreachable endpoint, etc.)
- **Stderr**: error messages explaining what is wrong (Saga displays these to the user)

This ensures all couriers are ready before any deliveries begin, avoiding partial delivery when one courier has misconfigured credentials.

### Info Subcommand

Couriers must support an `info` subcommand that outputs JSON describing the courier:

```bash
saga-courier-slack info
```

```json
{
  "name": "slack",
  "description": "Delivers tales to Slack channels via webhooks or the Slack API",
  "config_keys": [
    {
      "name": "channel",
      "description": "Slack channel to post to (e.g., #finance-updates)",
      "regex": "^#[a-z0-9_-]+$"
    },
    {
      "name": "icon_emoji",
      "description": "Emoji to use as the bot icon (e.g., :mega:)",
      "regex": "^(:[a-z0-9_+-]+:)?$"
    },
    {
      "name": "username",
      "description": "Display name for the bot"
    }
  ]
}
```

This allows `saga validate` to check that route config values match the courier's regex patterns (and infer which keys are required), and allows `saga couriers` to show helpful information.

### Output Contract (info)

- **Exit code 0**: info output successful, JSON on stdout
- **Exit code non-zero**: error (Saga treats this as a validation error for the courier)

### Credentials

Courier credentials (API tokens, webhook URLs, etc.) must NOT be stored in `config.jsonnet` since that file lives in git. Couriers should read credentials from:

1. **Environment variables** (preferred): e.g., `SAGA_COURIER_SLACK_TOKEN`
2. **A credentials file**: e.g., `~/.saga/credentials.json` (not in the repo)

The naming convention for environment variables should be `SAGA_COURIER_<NAME>_<KEY>`, for example:
- `SAGA_COURIER_SLACK_TOKEN`
- `SAGA_COURIER_WORDPRESS_API_KEY`
- `SAGA_COURIER_EMAIL_SMTP_PASSWORD`

Saga itself does not manage credentials — that is the courier's responsibility. Saga only passes the `courier.config` map from the route to the courier.

### Example: Minimal Courier in Bash

A courier can be as simple as a bash script:

```bash
#!/usr/bin/env bash
# saga-courier-slack

set -euo pipefail

case "${1:-}" in
  info)
    cat <<'EOF'
{
  "name": "slack",
  "description": "Posts tales to a Slack channel via webhook",
  "config_keys": [
    {
      "name": "channel",
      "description": "Slack channel to post to",
      "regex": "^#[a-z0-9_-]+$"
    }
  ]
}
EOF
    exit 0
    ;;
  validate-config)
    # Verify credentials are available
    : "${SAGA_COURIER_SLACK_WEBHOOK_URL:?Set SAGA_COURIER_SLACK_WEBHOOK_URL}"
    ;;
  tell)
    # Read JSON payload from stdin
    payload=$(cat)

    # Extract fields using jq
    channel=$(echo "$payload" | jq -r '.route.config.channel')
    chapter=$(echo "$payload" | jq -r '.chapter')
    tales=$(echo "$payload" | jq -r '.tales[] | "• *\(.summary)*\n\(.text)"')

    # Build Slack message
    message="*Chapter ${chapter}*\n\n${tales}"

    # Post to Slack
    webhook_url="${SAGA_COURIER_SLACK_WEBHOOK_URL:?Set SAGA_COURIER_SLACK_WEBHOOK_URL}"

    curl -sf -X POST "$webhook_url" \
      -H 'Content-Type: application/json' \
      -d "$(jq -n --arg channel "$channel" --arg text "$message" \
        '{channel: $channel, text: $text}')" \
      || { echo "Failed to post to Slack channel $channel" >&2; exit 1; }
    ;;
  *)
    echo "Usage: saga-courier-slack {info|validate-config|tell}" >&2
    exit 1
    ;;
esac
```

### Example: Courier in Go Using the Payload Package

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/saga-changelog/saga/pkg/courier"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "Usage: saga-courier-example {info|validate-config|tell}")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "info":
        info := courier.CourierInfo{
            Name:        "example",
            Description: "An example courier that prints tales to stdout",
            ConfigKeys: []courier.CourierConfigKey{
                {
                    Name:        "destination",
                    Description: "Output destination label",
                    Regex:       "^.+$",
                },
            },
        }
        json.NewEncoder(os.Stdout).Encode(info)

    case "validate-config":
        var payload courier.Payload
        if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
            fmt.Fprintf(os.Stderr, "failed to decode payload: %v\n", err)
            os.Exit(1)
        }
        // Verify credentials, connectivity, etc.
        dest := payload.Route.Config["destination"]
        if dest == "" {
            fmt.Fprintln(os.Stderr, "destination is empty")
            os.Exit(1)
        }

    case "tell":
        var payload courier.Payload
        if err := json.NewDecoder(os.Stdin).Decode(&payload); err != nil {
            fmt.Fprintf(os.Stderr, "failed to decode payload: %v\n", err)
            os.Exit(1)
        }

        dest := payload.Route.Config["destination"]
        for _, tale := range payload.Tales {
            fmt.Printf("[%s] %s: %s\n", dest, tale.Summary, tale.Text)
        }

    default:
        fmt.Fprintln(os.Stderr, "Usage: saga-courier-example {info|validate-config|tell}")
        os.Exit(1)
    }
}
```

---

## Validation

Validation is a critical part of Saga. It ensures that configuration is correct, feats reference valid audiences, and routes are properly wired before anything gets told.

### What Gets Validated

#### Configuration (`.saga/config.jsonnet`)

- File evaluates as valid Jsonnet
- Deserializes into the expected Go struct shape
- All route `audience` fields reference defined audience names
- All route `courier.name` fields reference an installed courier binary (error if not found)
- All route `name` fields are unique
- All courier binaries respond successfully to the `info` subcommand (error if they fail)
- Config values in each route match their courier's regex patterns (as reported by `info`); keys whose regex does not match an empty string are required

#### Feats (`.saga/feats/**/*.jsonnet`)

- File evaluates as valid Jsonnet
- Deserializes into the expected Go struct shape
- `summary` is a non-empty string
- `tales` array has at least one entry
- Every tale has a non-empty `audience` that references a defined audience name from config
- Every tale has non-empty `text`
- No duplicate `audience` values within a single feat's `tales`

### CI Integration

`saga validate` is designed to run in CI pipelines. Typical usage in a GitHub Action:

```yaml
- name: Validate Saga
  run: saga validate
```

This can also be used as a pre-commit hook or PR check to ensure that:
- New feats reference valid audiences
- Configuration changes don't break existing feats
- No dangling audience references exist

---

## Workflows

### Developer Workflow (Writing Feats)

1. Developer completes a feature or change
2. Developer runs `saga audiences` to see available audiences, their interests, and tone guidance
3. Developer creates a new `.jsonnet` file in `.saga/feats/pending/` with a descriptive slug filename
4. Developer writes the feat with a summary and one or more tales targeting specific audiences
5. Developer runs `saga validate` to check their feat (or CI does this on the PR)
6. Feat file is committed alongside the code changes in the same PR

### Release Workflow (Creating Chapters)

1. Release manager decides it's time for a new version
2. Release manager runs `saga pending` to review what feats are ready
3. Release manager runs `saga chapters create <version>` to bundle feats (optionally specifying specific feats to include)
4. The moved files are committed as part of the release commit (alongside version bumps, etc.)
5. The release is tagged, built, and prepared for deployment

### Deployment Workflow (Telling)

1. The new version is deployed to production
2. After deployment is confirmed successful, the deployer runs `saga tell <version> --dry-run` to preview
3. The deployer runs `saga tell <version>` to dispatch couriers
4. Saga reports success/failure per route

### Staged Rollout Workflow

1. After deployment, tell internal audiences first: `saga tell 2.4.0 --audience billing --audience engineering`
2. Verify internal notifications look correct
3. Tell external audiences: `saga tell 2.4.0 --audience partners`

---

## Key Dependencies

| Dependency | Import Path | Purpose |
|-----------|-------------|---------|
| Kong | `github.com/alecthomas/kong` | Struct-based CLI framework |
| go-jsonnet | `github.com/google/go-jsonnet` | Jsonnet evaluation, pure Go |

All other functionality uses the Go standard library (JSON, file I/O, `os/exec`, etc.).

---

## Future Considerations

These are ideas discussed during brainstorming that are not in scope for the initial version but should be kept in mind for the architecture.

### git integration

Should `saga chapters create <version>` commit the moved feats to git? Should it only work on a pristine repo? Is scripting this part of saga, or the tooling that calls saga?

### Smart sorting of chapters

Split chapters on `.` and `-`, then sort them first based on numeric interpretation, then alphabetically. Perhaps chapter format needs to be defined in config.jsonnet and validated by `saga validate`.

### Changelog Generation

A `saga chronicle` command could generate a traditional `CHANGELOG.md` from all chapters, providing backward compatibility with the ecosystem that expects one. It should accept `--audience` (defaulting to all audiences), and for each feat, pick the tale for that audience. Feats without a tale for the selected audience are skipped.

### Linkback Metadata

Feats could include a `links` field for referencing related URLs (documentation, migration guides) that couriers can include in their output.

### Amending Chapters

`saga chapters amend <version>` could move additional pending feats into an existing chapter, for cases where a feat was forgotten after chapter creation.

### Told State Tracking

A `.told.json` file per chapter could record success/failure per route when `saga tell` is run. This would enable:
- `saga chapters` showing told/untold/partially-told status per chapter
- Idempotent re-runs of `saga tell` that skip routes that already succeeded
- A clear audit trail of when and how tales were delivered

### Chapter Immutability Check

A `saga` subcommand that verifies no chaptered feats have been modified, for use in CI pipelines. It would use git to detect whether any files in `.saga/feats/chapters/` have been changed relative to a base ref. This enforces the rule that once feats are chaptered, they are set in stone and must not be edited.

### Feat Scaffolding

A `saga feats new <slug>` command could generate a skeleton `.jsonnet` file in `.saga/feats/pending/` with the available audiences pre-filled from config. This would reduce friction for developers unfamiliar with Jsonnet.

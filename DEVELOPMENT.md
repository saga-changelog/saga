# Development

Internal architecture and design decisions for saga contributors.

## Two-tier validation

Route configuration is validated at two separate tiers with different guarantees.

### Tier 1: structural validation (`saga validate`)

Runs in-process without credentials or network access. Saga calls each referenced courier's `info` subcommand to obtain its `courier.Info`, which declares `ConfigKeys` with optional regex patterns. Saga compiles those regexes and matches them against the route's config values. A regex that does not match the empty string implies the key is required.

`validate-route` is **not** invoked here. The courier binary is only called for `info`, and `info` results are cached per courier name so a courier referenced by multiple routes is invoked exactly once (`internal/cli/validate.go:validateCouriers`).

### Tier 2: operational validation (`saga tell`)

Before dispatching any tales, `saga tell` invokes each courier's `validate-route` subcommand with the `courier.Route` JSON on stdin. This checks operational readiness: credentials present in the environment, endpoints reachable, tokens not expired. Runs once per unique qualified route name (audience/route), not per tale.

Couriers are expected to be defensive in their `ValidateRoute` implementation. The Slack courier re-validates the channel regex even though saga already checked it in tier 1, because courier binaries can be invoked directly outside of saga and cannot assume prior validation. This overlap is by design.

### Where each tier runs in the codebase

| Tier | Entry point | What it checks |
|------|-------------|----------------|
| Structural | `cli.ValidateCmd.Run` → `validateCouriers` | Config key presence, regex match, audience references, feat structure |
| Operational | `cli.TellCmd.Run` → `dispatch.ValidateRoute` | Credentials, connectivity, platform-specific preconditions |

## Courier environment isolation

When saga forks a courier binary via `os/exec`, it constructs a filtered copy of the process environment (`internal/dispatch/dispatch.go:envForCourier`). All `SAGA_COURIER_*` variables are stripped except those matching the courier's own prefix. The prefix is derived as `SAGA_COURIER_<UPPER_NAME>__` where hyphens become underscores and the boundary is a **double underscore**. This prevents secret leakage between couriers and disambiguates couriers whose names share a prefix (`basecamp-messageboard` vs `basecamp-campfire` → `SAGA_COURIER_BASECAMP_MESSAGEBOARD__` vs `SAGA_COURIER_BASECAMP_CAMPFIRE__`).

Non-`SAGA_COURIER_*` environment variables (PATH, HOME, etc.) pass through unmodified.

## Tale text pipeline

Tale text passes through three stages, crossing a process boundary between parsing and rendering.

### 1. Authoring

Developers write CommonMark in the `text` field of each tale inside `.jsonnet` feat files. Jsonnet's `|||` block string syntax provides natural multiline support.

### 2. Parsing (`internal/taletext`)

`taletext.Parse` feeds the markdown source to goldmark (default CommonMark parser, no extensions) and walks the resulting AST, rejecting any construct outside the allowed subset:

- **Blocks**: paragraphs and level-2 headings only. No lists, blockquotes, code blocks, thematic breaks, or other heading levels.
- **Inlines**: plain text, bold (`**`), italic (`_`/`*`), and links. No code spans, images, or HTML.

The AST is flattened into `courier.TaleText`: a `[]Block`, each containing a `[]Inline`. Inlines are flat runs with a single `InlineStyle` and optional URL. Adjacent text fragments emitted by goldmark are merged. When styles nest (e.g. a bold link), **link takes precedence** because it carries a URL that would otherwise be lost. This is a lossy but intentional simplification: couriers receive a uniform structure they can render with a single switch on `InlineStyle`, without recursion.

Soft and hard line breaks within a paragraph collapse to a space. Paragraph breaks (blank lines in source) produce separate `Block` entries.

### 3. Rendering (courier-side)

Each courier converts `TaleText` to its platform's native markup. This happens in the courier binary's process, after JSON deserialization of the `Payload`. The courier never parses markdown -- it receives pre-parsed structured data. Examples:

| Courier | Output format | Bold | Link |
|---------|--------------|------|------|
| slack-legacy | mrkdwn | `*text*` | `<url\|text>` |
| basecamp-messageboard | HTML | `<strong>text</strong>` | `<a href="url">text</a>` |
| local-stdout | plain text with markers | `**text**` | `[text](url)` |

This separation means adding a new courier never requires touching the parser, and tightening the allowed markdown subset automatically applies to all couriers.

## Jsonnet sandboxing

Feat and config files are untrusted input (they arrive via pull requests). The jsonnet VM created by `internal/safejsonnet.MakeVM` is locked down:

- **Import path restriction**: a custom `Importer` resolves all import paths relative to the importing file's directory, then rejects any resolved path that does not have the `.saga/` directory as a prefix. Path traversal via `../` is normalized by `filepath.Clean` before the check.
- **Stack depth cap**: `MaxStack = 200`, preventing runaway recursion from consuming host resources.
- **No native functions**: the VM has zero registered `NativeFunction` callbacks, so evaluated jsonnet cannot invoke Go code or perform I/O beyond what the importer allows.

The VM is stateless with respect to evaluation: `EvaluateAnonymousSnippet` parses and evaluates independently each time. A single VM instance is reused across multiple files within `feat.LoadDir` for efficiency. A fresh VM is created per `LoadDir`/`LoadFile` call, not shared across commands.

## Workspace and module layout

The repository is a Go workspace (`go.work`) containing four modules:

```
go.work
├── .                                    saga CLI (main module)
├── couriers/saga-courier-local-stdout            separate go.mod
├── couriers/saga-courier-slack-legacy            separate go.mod
└── couriers/saga-courier-basecamp-messageboard   separate go.mod
```

Courier modules declare `replace github.com/saga-changelog/saga => ../..` to depend on the local `pkg/courier` package. The workspace ensures `go build ./...` at the root compiles everything, but each courier ships as an independent binary with its own dependency tree. The Basecamp courier pulls in `github.com/basecamp/basecamp-sdk/go`; that dependency does not affect the main saga binary or the other couriers.

`pkg/courier` is the only public Go package. Everything under `internal/` is inaccessible to courier authors or other importers.

## Courier plugin contract

Courier binaries are named `saga-courier-<name>` and discovered by scanning `PATH`. They must handle three subcommands:

| Subcommand | Stdin | Stdout | Exit 0 | Exit non-zero |
|---|---|---|---|---|
| `info` | -- | `courier.Info` JSON | success | saga treats as validation error |
| `validate-route` | `courier.Route` JSON | -- | route is operationally ready | stderr shown to user |
| `tell` | `courier.Payload` JSON | informational (prefixed and forwarded) | delivery succeeded | stderr captured, delivery marked failed |

For `tell`, saga prefixes both stdout and stderr lines with the courier name before forwarding to the user's terminal, so courier output is clearly attributed. Stderr is also buffered and included in the error message if the courier exits non-zero.

`pkg/courier.Run` is the entrypoint helper for Go-based couriers. It handles `os.Args` dispatch, JSON decoding from stdin, `signal.NotifyContext` for SIGINT/SIGTERM propagation, and exit codes. A courier author implements `courier.Courier` (three methods: `Info`, `ValidateRoute`, `Tell`) and calls `courier.Run(myCourier{})` from `main`.

## Config structure

`.saga/config.jsonnet` defines audiences with routes nested inline:

```jsonnet
{
  audiences: [
    {
      name: "engineering",
      // ...
      routes: [
        { name: "slack-legacy", courier: { name: "slack-legacy", config: { channel: "#eng" } } },
      ],
    },
  ],
}
```

Routes belong to their audience -- there is no separate top-level routes array and no `route.audience` back-reference in the config. When saga needs to iterate across all routes regardless of audience (for validation, dispatch), `config.AllRoutes()` flattens the tree into `[]AudienceRoute` pairs. `AudienceRoute.QualifiedName()` returns `"audience/route"` for use as a map key and in user-facing output.

## `saga tell` dispatch model

`saga tell` delivers one tale per courier invocation. For a chapter with N feats and a route whose audience has tales in M of those feats, saga forks the courier binary M times for that route. Each invocation receives a `courier.Payload` containing a single `courier.Tale`. This means each invocation produces one message on the destination platform.

Deliveries are dispatched sequentially. If a delivery fails, the error is recorded and remaining deliveries continue. The command exits non-zero if any delivery failed, reporting which ones.

# Saga

Saga is a changelog tool that replaces the traditional append-only `CHANGELOG.md` with a structured, file-per-change workflow, and then actually delivers those changes to the people who care about them.

Instead of writing one blob of release notes for everyone, a developer describes each change once, with different wordings for different audiences. Internal engineers see the API details and ticket numbers; customers see a friendly summary on the company message board; finance sees the billing-relevant bits in their Slack channel. Saga routes the right wording to the right place at the right time, so the same change can land in all of them without anyone copy-pasting between tools.

## How it fits together

Each change is captured as a **feat**: a small Jsonnet file checked into `.saga/feats/pending/` alongside the code that implements the change. A feat contains one or more **tales**, which are audience-specific descriptions, each with a short title and markdown text. A developer writing a feat picks which audiences need to know about the change and writes a tale for each of them, using `saga audiences` to remind themselves of who those audiences are and what tone to use.

When it's time to cut a release, `saga chapters create <version>` bundles every pending feat into a **chapter**: a versioned directory under `.saga/feats/chapters/`. The feats move from pending into the chapter and are committed as part of the release. Chapters are immutable; once a feat is chaptered, it stays exactly as written.

After the release is deployed, `saga tell <version>` dispatches each tale to its destinations. The wiring from audience to destination lives in `.saga/config.jsonnet` as **routes**: each audience declares one or more routes, and each route names a **courier** (like `slack-legacy` or `basecamp-messageboard`) with the configuration that courier needs (a channel name, a project ID, and so on). For every tale in the chapter, saga looks up the matching audience's routes and invokes the courier once per tale, passing it a structured payload. The courier is a standalone binary on your PATH named `saga-courier-<name>` that turns that payload into a real message on its platform.

Saga itself doesn't know anything about Slack or Basecamp or any other destination. Couriers are external binaries that implement a tiny contract, which keeps saga small and lets you add a new destination without touching saga's source.

## Getting started

Install saga and whichever couriers you need:

```sh
go install github.com/saga-changelog/saga@latest
go install github.com/saga-changelog/saga/couriers/saga-courier-slack-legacy@latest
go install github.com/saga-changelog/saga/couriers/saga-courier-basecamp-messageboard@latest
```

Inside your project's git repository, initialize saga:

```sh
saga init
```

Edit `.saga/config.jsonnet` to define your audiences and their routes. Then, as you work on features and fixes, drop `.jsonnet` feat files into `.saga/feats/pending/`, one per change. `saga audiences` shows the audiences you've defined with their tone and interest guidance so you know what to write. `saga validate` checks everything is well-formed (run it in CI to keep things honest). `saga pending` shows what's accumulated since the last release.

When you're ready to ship:

```sh
saga chapters create v1.2.3       # bundle pending feats into a chapter
git commit -am "release v1.2.3"   # commit the move
# ... deploy ...
saga tell v1.2.3                  # dispatch to all couriers
```

You can preview with `saga tell v1.2.3 --dry-run`, or scope a run to a single audience with `--audience engineering` (and further to specific routes within that audience with `--routes slack`).

Credentials for couriers come from environment variables named `SAGA_COURIER_<NAME>__<KEY>` (note the double underscore between the courier name and the key). Saga filters these per courier so each one only sees its own secrets.

## Couriers

- [`saga-courier-basecamp-messageboard`](./couriers/saga-courier-basecamp-messageboard/README.md): posts tales to a Basecamp message board
- [`saga-courier-slack-app`](./couriers/saga-courier-slack-app/README.md): delivers tales to Slack channels via the Web API with Block Kit formatting
- [`saga-courier-slack-legacy`](./couriers/saga-courier-slack-legacy/README.md): delivers tales to Slack channels via incoming webhooks
- [`saga-courier-local-file`](./couriers/saga-courier-local-file/README.md): appends tales to a local file in markdown format
- [`saga-courier-local-stdout`](./couriers/saga-courier-local-stdout/README.md): prints tales to stdout, useful for testing and local development

Writing your own courier is straightforward: implement the `Courier` interface in `github.com/saga-changelog/saga/pkg/courier` and call `courier.Run` from `main`. The package handles argument parsing, payload decoding, signal-aware context, and exit codes, so your code only needs to describe the courier and deliver the tale.

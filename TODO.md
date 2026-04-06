# TODO

## Shortlist

Things I want to build now.

### Markdown tales:

- Support monospace (backticks).
- Support lists.

- Should tales have "read more" section? For longer texts, that are then posted differently depending on the courier? For example in slack the "read more" section could be posted in the thread?
  - Horizontal rule (divider in basecamp, comments in thread in slack)
  - We need to switch slack to web api with bot token instead of webhook url.

- Probably try autolinks support in markdown parser; https://github.github.com/gfm/#autolinks-extension-

### Repo health

- Improve linting, research options (golangci-lint is sloooowww) and configsets.
- Implement github actions.

### Publish

- Cross-platform builds for saga and all in-repo couriers published via CD to github

### Theming

Align CLI theme with thematic lore surrounding 'saga'. Ancient / scrolls / papyrus / rock / bronze / iron / dirt / gold / bloodred / sun / moon / nordic / greek / pagan.

### Renaming of couriers

Platforms often have different ways to post, and may move to newer APIs. We're already encountering that with slack, the current implementation is based on the "incoming webhook", but to use more message markup features we should use the newer bot integration. Hence the current courier should be renamed to `saga-courier-slack-webhook` and the new one could be `saga-courier-slack-bot`. For basecamp it's actually `basecamp-messageboard` and other integrations could be `basecamp-todo` and `basecamp-campfire`, providing users with different options of integrating saga into their workflow/process/communications.

## Future Considerations

Ideas discussed during brainstorming that are not in scope for the initial version but should be kept in mind for the architecture.

### git integration

Should `saga chapters create <version>` commit the moved feats to git? Should it only work on a pristine repo? Is scripting this part of saga, or the tooling that calls saga?

### Chronicle / Changelog Generation

A `saga chronicle` command could generate a traditional `CHANGELOG.md`-like file from all chapters, providing backward compatibility with the ecosystem that expects one. It should accept a mandatory `--audience`, and for each feat, pick the tale for that audience. Feats without a tale for the selected audience are skipped.

### Amending Chapters

_I dont think we should do this_

`saga chapters amend <version>` could move additional pending feats into an existing chapter, for cases where a feat was forgotten after chapter creation.

### Chapter State Tracking

Add a metadata file to chapters that track when the chapter was created.

### Told State Tracking

A `.told.json` file per chapter could record success/failure per route when `saga tell` is run. This would enable:
- `saga chapters` showing told/untold/partially-told status per chapter
- Idempotent re-runs of `saga tell` that skip routes that already succeeded
- A clear audit trail of when and how tales were delivered

### Chapter Immutability Check

A `saga ci` subcommand that verifies no chaptered feats have been modified, for use in CI pipelines. It would use git to detect whether any files in `.saga/feats/chapters/` have been changed relative to a base ref. This enforces the rule that once feats are chaptered, they are set in stone and must not be edited.

### Feat Scaffolding

A `saga feats new <slug>` command could generate a skeleton `.jsonnet` file in `.saga/feats/pending/` with the available audiences pre-filled from config. This would reduce friction for developers unfamiliar with Jsonnet.

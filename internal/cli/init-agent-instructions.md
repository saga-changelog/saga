
## Saga changelog

This project uses [saga](https://github.com/saga-changelog/saga) to communicate software changes to the right people through the right channels.

When you make a change that is notable to one or more audiences, create a feat file in `.saga/feats/pending/`. A feat is a jsonnet file containing tales — audience-specific descriptions of the change.

Before writing a feat, read `.saga/config.jsonnet` to see which audiences are defined, what they care about, and what tone to use. Write a tale for each audience that would find the change relevant. Not every change needs a tale for every audience.

Feat file structure (`.saga/feats/pending/<slug>.jsonnet`):

```jsonnet
{
  tales: [
    {
      audience: '<audience-name>',
      title: '<single-line title>',
      text: |||
        <markdown body>
      |||,
    },
  ],
}
```

Rules:
- The slug should be a short kebab-case name for the change (e.g. `api-pagination`, `fix-upload-timeout`)
- Titles are plain text, single line, no period at the end
- Text supports a limited markdown subset: paragraphs, level-2 headings, **bold**, _italic_, and [links](url). No lists, code blocks, or images.
- Match the tone specified for each audience in the config
- Do not create feats for trivial changes (typo fixes, internal refactors with no user-facing impact)

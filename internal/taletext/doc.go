// Package taletext parses the author-facing markdown of a tale's text
// into the neutral courier.TaleText structure sent to courier binaries.
//
// Only a small, safe subset of CommonMark is permitted:
//
//   - Paragraphs
//   - Level-2 headings (## Text)
//   - **bold** and *italic* (plus _italic_)
//   - Links: [label](url) and autolinks <https://...>
//   - Inline code spans: `code`
//   - Fenced code blocks: ```language ... ```
//
// Anything else (lists, block quotes, images, tables, raw HTML,
// thematic breaks, or headings of any other level) is a parse error.
// The intent is to keep tale content portable across Slack, Basecamp,
// and any future destinations without pulling a markdown parser into
// each courier binary.
//
// # Flat inline model
//
// Tale text is modelled as a list of blocks each containing a flat list
// of styled inline runs (see courier.TaleText). There is no nesting of
// styles. When the source has nested emphasis (e.g. a bold-inside-link
// or link-inside-bold), the link style is preferred over bold/italic
// because links carry a URL that the other styles do not. Emphasis
// nested inside other emphasis resolves to the innermost style.
package taletext

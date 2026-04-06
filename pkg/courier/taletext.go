package courier

// TaleText is the parsed, structured representation of a tale's text.
// It is a minimal document model: a list of block-level elements, each
// containing a flat list of styled inline runs. There is no nesting of
// styles: if a link label was also bold in the source, only the link
// style is preserved (the link carries the URL that the bold would not).
type TaleText struct {
	// Blocks are the block-level elements of the text, in document order.
	Blocks []Block `json:"blocks"`
}

// Block is a block-level element within a TaleText.
type Block struct {
	// Kind is the block's type (paragraph, heading, ...).
	Kind BlockKind `json:"kind"`

	// Inlines is the flat list of styled text runs that make up the block.
	Inlines []Inline `json:"inlines"`
}

// BlockKind identifies the type of a Block.
type BlockKind string

const (
	// BlockParagraph is a paragraph of inline runs.
	BlockParagraph BlockKind = "paragraph"
	// BlockHeading is always a level-2 heading; saga only permits a
	// single heading level in tale text.
	BlockHeading BlockKind = "heading"
)

// Inline is a single styled text run within a Block.
type Inline struct {
	// Style is the rendering style for this run. An empty value means
	// plain (unstyled) text.
	Style InlineStyle `json:"style"`

	// Text is the literal text of the run. For links, this is the
	// visible label.
	Text string `json:"text"`

	// URL is set only when Style == InlineLink.
	URL string `json:"url,omitempty"`
}

// InlineStyle identifies the rendering style of an Inline.
type InlineStyle string

const (
	// InlinePlain is unstyled text.
	InlinePlain InlineStyle = ""
	// InlineBold is bold text.
	InlineBold InlineStyle = "bold"
	// InlineItalic is italic text.
	InlineItalic InlineStyle = "italic"
	// InlineLink is a hyperlink; the Inline's URL field carries the target.
	InlineLink InlineStyle = "link"
)

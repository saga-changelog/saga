package taletext

import (
	"fmt"
	"strings"

	"github.com/saga-changelog/saga/pkg/courier"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Parse parses a tale text string and returns its structured form. It
// returns an error the moment it encounters any construct outside the
// allowed subset, naming the offending construct and its line number.
func Parse(source string) (courier.TaleText, error) {
	src := []byte(source)
	md := goldmark.New() // default CommonMark, no extensions
	root := md.Parser().Parse(text.NewReader(src))

	var tt courier.TaleText
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		block, err := parseBlock(n, src)
		if err != nil {
			return courier.TaleText{}, err
		}
		tt.Blocks = append(tt.Blocks, block)
	}
	return tt, nil
}

// parseBlock converts a goldmark block-level node into a courier.Block,
// rejecting any kind that is not explicitly allowed.
func parseBlock(n ast.Node, src []byte) (courier.Block, error) {
	switch n.Kind() {
	case ast.KindParagraph:
		inlines, err := parseInlines(n, src)
		if err != nil {
			return courier.Block{}, err
		}
		return courier.Block{Kind: courier.BlockParagraph, Inlines: inlines}, nil

	case ast.KindHeading:
		h := n.(*ast.Heading)
		if h.Level != 2 {
			return courier.Block{}, fmt.Errorf("%sonly level-2 headings (##) are allowed, got level %d", lineHint(n, src), h.Level)
		}
		inlines, err := parseInlines(n, src)
		if err != nil {
			return courier.Block{}, err
		}
		return courier.Block{Kind: courier.BlockHeading, Inlines: inlines}, nil

	case ast.KindFencedCodeBlock:
		cb := n.(*ast.FencedCodeBlock)
		var content strings.Builder
		for i := 0; i < cb.Lines().Len(); i++ {
			seg := cb.Lines().At(i)
			content.Write(seg.Value(src))
		}
		text := strings.TrimRight(content.String(), "\n")
		info := ""
		if cb.Info != nil {
			info = string(cb.Info.Text(src))
		}
		return courier.Block{Kind: courier.BlockCodeBlock, Text: text, Info: info}, nil

	default:
		return courier.Block{}, fmt.Errorf("%sdisallowed block-level construct %q", lineHint(n, src), n.Kind().String())
	}
}

// parseInlines walks the inline children of a block (paragraph or
// heading) and flattens them into a list of styled runs.
func parseInlines(parent ast.Node, src []byte) ([]courier.Inline, error) {
	var out []courier.Inline
	if err := walkInlines(parent, src, courier.InlinePlain, "", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// walkInlines recursively visits the inline descendants of n, preserving
// the innermost styling context. Link style takes precedence over bold
// and italic, because a link carries a URL that would otherwise be lost.
func walkInlines(n ast.Node, src []byte, style courier.InlineStyle, url string, out *[]courier.Inline) error {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {

		case ast.KindText:
			t := c.(*ast.Text)
			if txt := string(t.Segment.Value(src)); txt != "" {
				appendInline(out, style, url, txt)
			}
			// Soft and hard line breaks both collapse to a space.
			// We don't model explicit line breaks in the flat inline
			// representation; authors use blank lines for paragraph
			// breaks instead.
			if t.SoftLineBreak() || t.HardLineBreak() {
				appendInline(out, style, url, " ")
			}

		case ast.KindString:
			s := c.(*ast.String)
			if txt := string(s.Value); txt != "" {
				appendInline(out, style, url, txt)
			}

		case ast.KindEmphasis:
			e := c.(*ast.Emphasis)
			var nested courier.InlineStyle
			switch e.Level {
			case 1:
				nested = courier.InlineItalic
			case 2:
				nested = courier.InlineBold
			default:
				return fmt.Errorf("%sunsupported emphasis level %d", lineHint(c, src), e.Level)
			}
			// Link style outranks emphasis (URL > visual style). If
			// we're already inside a link, keep the link context;
			// otherwise switch into the emphasis style.
			if style == courier.InlineLink {
				if err := walkInlines(c, src, style, url, out); err != nil {
					return err
				}
			} else {
				if err := walkInlines(c, src, nested, "", out); err != nil {
					return err
				}
			}

		case ast.KindLink:
			l := c.(*ast.Link)
			if err := walkInlines(c, src, courier.InlineLink, string(l.Destination), out); err != nil {
				return err
			}

		case ast.KindAutoLink:
			a := c.(*ast.AutoLink)
			u := string(a.URL(src))
			appendInline(out, courier.InlineLink, u, u)

		case ast.KindCodeSpan:
			// Code spans contain literal text segments; extract
			// and emit as a single code-styled inline run.
			var cb strings.Builder
			for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
				if gc.Kind() == ast.KindText {
					t := gc.(*ast.Text)
					cb.Write(t.Segment.Value(src))
					if t.SoftLineBreak() || t.HardLineBreak() {
						cb.WriteByte(' ')
					}
				}
			}
			if txt := cb.String(); txt != "" {
				appendInline(out, courier.InlineCode, "", txt)
			}

		default:
			return fmt.Errorf("%sdisallowed inline construct %q", lineHint(c, src), c.Kind().String())
		}
	}
	return nil
}

// appendInline appends a new inline run, merging it with the previous
// run when both styles and URLs match. This keeps the inline list tidy
// when the parser emits adjacent text fragments.
func appendInline(out *[]courier.Inline, style courier.InlineStyle, url, text string) {
	if n := len(*out); n > 0 {
		last := &(*out)[n-1]
		if last.Style == style && last.URL == url {
			last.Text += text
			return
		}
	}
	*out = append(*out, courier.Inline{Style: style, Text: text, URL: url})
}

// lineHint returns a short "line N: " prefix for validation errors when
// goldmark has recorded source segment information for the node. Inline
// nodes don't carry line information (and goldmark panics if you ask), so
// we walk up to the nearest block ancestor before querying.
func lineHint(n ast.Node, src []byte) string {
	for n != nil && n.Type() != ast.TypeBlock {
		n = n.Parent()
	}
	if n == nil {
		return ""
	}
	lines := n.Lines()
	if lines == nil || lines.Len() == 0 {
		return ""
	}
	start := lines.At(0).Start
	if start < 0 || start > len(src) {
		return ""
	}
	// Line numbers are 1-indexed; count newlines preceding the offset.
	line := 1
	for _, b := range src[:start] {
		if b == '\n' {
			line++
		}
	}
	return fmt.Sprintf("line %d: ", line)
}

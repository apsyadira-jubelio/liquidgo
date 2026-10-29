package tags

import (
	"regexp"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

var rawSyntax = regexp.MustCompile(`^\s*$`)

// RawTag represents a raw tag that outputs Liquid code as text.
type RawTag struct {
	*liquid.Block
	body string
}

// NewRawTag creates a new RawTag.
func NewRawTag(tagName, markup string, parseContext liquid.ParseContextInterface) (*RawTag, error) {
	// Validate markup is empty
	if !rawSyntax.MatchString(markup) {
		var locale *liquid.I18n
		if pc, ok := parseContext.(*liquid.ParseContext); ok {
			locale = pc.Locale()
			msg := locale.Translate("errors.syntax.tag_unexpected_args", map[string]interface{}{"tag": tagName})
			return nil, liquid.NewSyntaxError(msg)
		}
		return nil, liquid.NewSyntaxError("Liquid syntax error: tag unexpected args")
	}

	block := liquid.NewBlock(tagName, markup, parseContext)
	block.SetBlockDelimiter("endraw")

	return &RawTag{
		Block: block,
		body:  "",
	}, nil
}

// rawEndTag mirrors Ruby's Raw::FullTokenPossiblyInvalid:
// /\A(.*)#{TagStart}#{WhitespaceControl}?\s*(\w+)\s*(.*)?#{WhitespaceControl}?#{TagEnd}\z/om
var rawEndTag = regexp.MustCompile(`(?s)\A(.*)\{%-?\s*(\w+)\s*(.*)?-?%\}\z`)

// Parse parses the raw tag body. As in Ruby, the body is kept verbatim and
// `-%}` on the end tag trims the text that follows.
func (r *RawTag) Parse(tokenizer *liquid.Tokenizer) error {
	var body strings.Builder
	for {
		token := tokenizer.Shift()
		if token == "" {
			r.body = body.String()
			return liquid.NewSyntaxError("Liquid syntax error: " + r.BlockName() + " tag was never closed")
		}
		if m := rawEndTag.FindStringSubmatch(token); m != nil && m[2] == r.BlockDelimiter() {
			r.ParseContext().SetTrimWhitespace(len(token) >= 3 && token[len(token)-3] == '-')
			body.WriteString(m[1])
			r.body = body.String()
			return nil
		}
		body.WriteString(token)
	}
}

// RenderToOutputBuffer renders the raw tag.
func (r *RawTag) RenderToOutputBuffer(context liquid.TagContext, output *strings.Builder) {
	output.WriteString(r.body)
}

// Nodelist returns the nodelist (just the body as a string).
func (r *RawTag) Nodelist() []interface{} {
	return []interface{}{r.body}
}

// Blank returns true if body is empty.
func (r *RawTag) Blank() bool {
	return r.body == ""
}

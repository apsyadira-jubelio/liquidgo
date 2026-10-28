package tags

import (
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

// BreakTag represents a break tag that stops a for loop from iterating.
type BreakTag struct {
	*liquid.Tag
}

// NewBreakTag creates a new BreakTag.
func NewBreakTag(tagName, markup string, parseContext liquid.ParseContextInterface) *BreakTag {
	return &BreakTag{
		Tag: liquid.NewTag(tagName, markup, parseContext),
	}
}

// RenderToOutputBuffer renders the break tag by pushing a break interrupt.
func (b *BreakTag) RenderToOutputBuffer(context liquid.TagContext, output *strings.Builder) {
	interrupt := liquid.NewBreakInterrupt()
	context.PushInterrupt(interrupt)
}

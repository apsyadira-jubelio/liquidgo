package tags

import (
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
)

// EchoTag represents an echo tag that outputs an expression.
type EchoTag struct {
	*liquid.Tag
	variable *liquid.Variable
}

// NewEchoTag creates a new EchoTag.
func NewEchoTag(tagName, markup string, parseContext liquid.ParseContextInterface) *EchoTag {
	return &EchoTag{
		Tag:      liquid.NewTag(tagName, markup, parseContext),
		variable: liquid.NewVariable(markup, parseContext),
	}
}

// RenderToOutputBuffer renders the echo tag by rendering the variable.
func (e *EchoTag) RenderToOutputBuffer(context liquid.TagContext, output *strings.Builder) {
	// Render the variable and append to output
	val := e.variable.Render(context)
	output.WriteString(liquid.ToS(val, nil))
}

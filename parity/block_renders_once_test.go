package parity

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

type plainBlock struct{ *liquid.Block }

// Ruby Block#render_to_output_buffer renders the body once. A custom block
// that only embeds *liquid.Block rendered it three times (override probing),
// so side effects such as increment ran thrice.
func TestCustomBlockRendersBodyOnce(t *testing.T) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.RegisterTag("plain", tags.TagConstructor(func(name, markup string, pc liquid.ParseContextInterface) (interface{}, error) {
		return &plainBlock{liquid.NewBlock(name, markup, pc)}, nil
	}))
	tmpl, err := liquid.ParseTemplate(`{% plain %}{% increment c %}{% endplain %}|{% increment c %}`, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	if got := tmpl.Render(nil, nil); got != "0|1" {
		t.Errorf("got %q, want %q", got, "0|1")
	}
}

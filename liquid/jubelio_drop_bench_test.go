package liquid_test

import (
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

type benchProductDrop struct {
	liquid.Drop
	Handle string
}

func (p *benchProductDrop) Title() string      { return "Shirt" }
func (p *benchProductDrop) ComparePrice() int  { return 1200 }
func (p *benchProductDrop) Available() bool    { return true }
func (p *benchProductDrop) FeaturedMedia() any { return nil }

// BenchmarkDropLookups hammers the drop/map lookup path: drop methods and
// fields, snake_case names, and map keys that miss (which fall through to the
// drop-method probe on the map's type).
func BenchmarkDropLookups(b *testing.B) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	src := `{% for i in (1..500) %}{{ product.title }}{{ product.compare_price }}{{ product.handle }}{% if product.available %}y{% endif %}{{ settings.missing_key }}{{ section.settings.heading }}{{ section.blocks.size }}{% endfor %}`
	tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		b.Fatal(err)
	}
	assigns := map[string]interface{}{
		"product":  &benchProductDrop{Handle: "shirt"},
		"settings": map[string]interface{}{"a": 1},
		"section": map[string]interface{}{
			"settings": map[string]interface{}{"heading": "Hi"},
			"blocks":   []interface{}{1, 2},
		},
	}
	b.ReportAllocs()
	for b.Loop() {
		tmpl.Render(assigns, nil)
	}
}

// BenchmarkFilterInvocation measures filter dispatch (method lookup per call).
func BenchmarkFilterInvocation(b *testing.B) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	src := `{% for i in (1..500) %}{{ title | upcase | append: "x" | strip_html | size | plus: 1 | at_least: 5 }}{% endfor %}`
	tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		b.Fatal(err)
	}
	assigns := map[string]interface{}{"title": "<b>Shirt</b>"}
	b.ReportAllocs()
	for b.Loop() {
		tmpl.Render(assigns, nil)
	}
}

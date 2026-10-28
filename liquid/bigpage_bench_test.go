package liquid_test

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// BenchmarkBigPage renders a ~337KB page from a 5000-iteration loop. With a
// string-concatenating output buffer this is quadratic in output size.
func BenchmarkBigPage(b *testing.B) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	src := `{% for i in (1..5000) %}<div class="product-card" data-index="{{ i }}">Item {{ i }} lorem</div>
{% endfor %}`
	tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	var n int
	for b.Loop() {
		n = len(tmpl.Render(map[string]interface{}{}, nil))
	}
	b.ReportMetric(float64(n), "out-bytes")
}

// BenchmarkParseSection parses a Dawn-like section repeatedly, so any
// cross-template parse cache would show up here.
func BenchmarkParseSection(b *testing.B) {
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	src := strings.Repeat(`{%- for block in section.blocks -%}
  <div class="{{ block.settings.class | escape }}" {{ block.shopify_attributes }}>
    {%- if block.settings.image != blank -%}{{ block.settings.image | image_url: width: 800 | image_tag }}{%- endif -%}
    <h2>{{ block.settings.heading | default: section.settings.title }}</h2>
    {%- assign p = product.variants.first.price | money -%}{{ p }} {{ settings.currency_code_enabled }}
  </div>
{%- endfor -%}
`, 40)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env}); err != nil {
			b.Fatal(err)
		}
	}
}

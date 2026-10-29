package parity

import "testing"

// Ruby Raw#parse ends on any `{%-? endraw -?%}` (FullTokenPossiblyInvalid),
// keeps the body verbatim, and lets `-%}` on endraw trim the following text.
func TestRawEndTagWhitespaceControl(t *testing.T) {
	cases := map[string]string{
		// raw_tag_test.rb test_output_in_raw
		"> {%- raw -%}{{ test }}{%- endraw -%} <":   ">{{ test }}<",
		" {%- raw -%}\n{{ test }}\n{%- endraw -%} ": "\n{{ test }}\n",
		"[{% raw %} {{ a }} {%- endraw -%}  ]":      "[ {{ a }} ]",
		"[{%- raw -%} {{ a }} {% endraw %} ]":       "[ {{ a }}  ]",
		"[{% raw %}{{ a }}{%endraw%}]":              "[{{ a }}]",
		"[{% raw %}{% if %}{%- endraw %}]":          "[{% if %}]",
		"  {%- raw %}x{% endraw -%}\n  [y]":         "x[y]",
	}
	for src, want := range cases {
		assertRender(t, src, want, nil)
	}
}

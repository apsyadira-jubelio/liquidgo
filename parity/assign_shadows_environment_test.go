package parity

import "testing"

// Ruby's Context#find_variable checks every scope (the outer scope included,
// which is where a top-level assign writes) before the environments. Checking
// environments first meant `{% assign pages = 1 %}` could not shadow a global.
func TestAssignShadowsEnvironmentVariable(t *testing.T) {
	data := map[string]interface{}{"pages": "global", "settings": map[string]interface{}{"a": 1}}
	assertRender(t, `{% assign pages = 1 %}{{ pages }}`, "1", data)
	assertRender(t, `{{ pages }}{% assign pages = 'local' %}{{ pages }}`, "globallocal", data)
	assertRender(t, `{% capture settings %}x{% endcapture %}{{ settings }}`, "x", data)
	assertRender(t, `{% for i in (1..2) %}{% assign pages = i %}{% endfor %}{{ pages }}`, "2", data)
}

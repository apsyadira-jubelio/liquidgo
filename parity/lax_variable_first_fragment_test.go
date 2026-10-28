package parity

import "testing"

// Ruby's lax Variable parse matches /(QuotedFragment)(.*)/: the name is the
// first fragment and anything after it that is not a filter is ignored, so
// `{{ a and b }}` renders a. The Go regex took everything up to the first
// pipe as the name, which parses to nil.
func TestLaxVariableTakesFirstFragment(t *testing.T) {
	data := map[string]interface{}{"a": "A", "b": "B"}
	assertRender(t, `{{ a and b }}`, "A", data)
	assertRender(t, `{{ a or b | downcase }}`, "a", data)
	assertRender(t, `{% assign x = a and b %}{{ x }}`, "A", data)
	assertRender(t, `{{ 'q t' extra }}`, "q t", data)
	assertRender(t, `{{ a | append: b }}`, "AB", data)
}

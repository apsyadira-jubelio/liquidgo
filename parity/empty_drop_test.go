package parity

import "testing"

type cartLike struct{ items []interface{} }

func (c cartLike) Empty() bool { return len(c.items) == 0 }
func (c cartLike) Blank() bool { return len(c.items) == 0 }

// Ruby Condition: `left.respond_to?(:empty?) ? left.empty? : …`. A drop that
// answers empty?/blank? (Shopify CartDrop) decides `== empty` / `== blank`.
func TestDropAnswersEmptyAndBlank(t *testing.T) {
	empty := map[string]interface{}{"cart": cartLike{}}
	full := map[string]interface{}{"cart": cartLike{items: []interface{}{1}}}
	assertRender(t, `{% if cart == empty %}y{% else %}n{% endif %}`, "y", empty)
	assertRender(t, `{% if cart != empty %}y{% else %}n{% endif %}`, "n", empty)
	assertRender(t, `{% if cart == empty %}y{% else %}n{% endif %}`, "n", full)
	assertRender(t, `{% if cart == blank %}y{% else %}n{% endif %}`, "y", empty)
	assertRender(t, `{% if cart == blank %}y{% else %}n{% endif %}`, "n", full)
}

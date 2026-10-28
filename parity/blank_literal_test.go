package parity

import "testing"

// Shopify's `blank` is ActiveSupport's blank?: true for nil, false, empty or
// whitespace-only strings, and empty arrays and hashes. Empty collections and
// false compared as not blank, so `{% if list != blank %}` rendered for [].
func TestBlankMatchesActiveSupportBlank(t *testing.T) {
	data := map[string]interface{}{
		"empty_list": []interface{}{},
		"typed_list": []string{},
		"empty_hash": map[string]interface{}{},
		"no":         false,
		"spaces":     "  \n",
		"list":       []interface{}{1},
		"hash":       map[string]interface{}{"a": 1},
		"zero":       0,
		"yes":        true,
		"word":       "x",
	}
	for _, name := range []string{"empty_list", "typed_list", "empty_hash", "no", "spaces", "nothing"} {
		assertRender(t, `{% if `+name+` == blank %}T{% else %}F{% endif %}`, "T", data)
		assertRender(t, `{% if `+name+` != blank %}T{% endif %}`, "", data)
	}
	for _, name := range []string{"list", "hash", "zero", "yes", "word"} {
		assertRender(t, `{% if `+name+` == blank %}T{% else %}F{% endif %}`, "F", data)
	}
}

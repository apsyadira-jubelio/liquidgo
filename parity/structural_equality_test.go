package parity

import "testing"

type eqDrop struct{ Name string }

// Ruby's == is structural for Hash and Array (with numeric elements compared
// across Integer/Float) and identity for drops. Go's == panicked comparing two
// maps or two slices, which Dawn hits with
// `variant == product.selected_or_first_available_variant`.
func TestEqualityIsStructuralForHashAndArray(t *testing.T) {
	shared := &eqDrop{Name: "a"}
	data := map[string]interface{}{
		"a":     map[string]interface{}{"id": 1, "tags": []interface{}{"x"}},
		"b":     map[string]interface{}{"id": 1.0, "tags": []interface{}{"x"}},
		"c":     map[string]interface{}{"id": 2},
		"l1":    []interface{}{1, "two"},
		"l2":    []interface{}{1, "two"},
		"l3":    []string{"one"},
		"l4":    []interface{}{"one"},
		"d1":    shared,
		"d2":    shared,
		"other": &eqDrop{Name: "a"},
	}
	assertRender(t, `{% if a == b %}T{% else %}F{% endif %}`, "T", data)
	assertRender(t, `{% if a == c %}T{% else %}F{% endif %}`, "F", data)
	assertRender(t, `{% if a != c %}T{% else %}F{% endif %}`, "T", data)
	assertRender(t, `{% if l1 == l2 %}T{% else %}F{% endif %}`, "T", data)
	assertRender(t, `{% if l3 == l4 %}T{% else %}F{% endif %}`, "T", data)
	assertRender(t, `{% if l1 == a %}T{% else %}F{% endif %}`, "F", data)
	assertRender(t, `{% if d1 == d2 %}T{% else %}F{% endif %}`, "T", data)
	assertRender(t, `{% if d1 == other %}T{% else %}F{% endif %}`, "F", data)
	assertRender(t, `{% case a %}{% when b %}MATCH{% else %}NO{% endcase %}`, "MATCH", data)
}

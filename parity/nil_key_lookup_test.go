package parity

import "testing"

// Ruby's hash[nil] is nil. A nil key into a map (e.g. linklists[nil] when the
// handle setting is unset) panicked in reflection.
func TestNilKeyLookupIsNil(t *testing.T) {
	data := map[string]interface{}{
		"linklists": map[string]interface{}{"main": "menu"},
		"typed":     map[string]string{"main": "menu"},
		"byid":      map[int]string{1: "one"},
	}
	assertRender(t, `[{{ linklists[section.settings.menu] }}]`, "[]", data)
	assertRender(t, `[{{ typed[nothing] }}]`, "[]", data)
	assertRender(t, `[{{ byid[nothing] }}]`, "[]", data)
	assertRender(t, `{{ typed["main"] }}{{ byid[1] }}`, "menuone", data)
}

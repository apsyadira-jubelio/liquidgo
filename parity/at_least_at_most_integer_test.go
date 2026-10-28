package parity

import "testing"

// Ruby's at_least/at_most return whichever Utils.to_number value wins, so
// Integer inputs stay Integer: `5000 | at_least: 1` is 5000, not 5000.0.
func TestAtLeastAtMostKeepInteger(t *testing.T) {
	data := map[string]interface{}{"n": 5000, "s": "7"}
	assertRender(t, `{{ n | at_least: 1 }}`, "5000", data)
	assertRender(t, `{{ n | at_least: 9000 }}`, "9000", data)
	assertRender(t, `{{ n | at_most: 10 }}`, "10", data)
	assertRender(t, `{{ s | at_most: 100 }}`, "7", data)
	assertRender(t, `{{ n | at_most: 1.5 }}`, "1.5", data)
	assertRender(t, `{{ 2.5 | at_least: 1 }}`, "2.5", data)
}

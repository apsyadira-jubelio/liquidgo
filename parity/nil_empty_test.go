package parity

import "testing"

// Ruby Condition#equal_variables: `right.respond_to?(:empty?) ?
// right.empty? : nil`. nil and numbers don't respond to empty?, so they never
// equal `empty`; only empty strings, arrays and hashes do.
func TestNilIsNotEmpty(t *testing.T) {
	data := map[string]interface{}{"s": "", "a": []interface{}{}, "n": 0}
	cases := map[string]string{
		`{% if nothing == empty %}y{% else %}n{% endif %}`: "n",
		`{% if nothing != empty %}y{% else %}n{% endif %}`: "y",
		`{% if n == empty %}y{% else %}n{% endif %}`:       "n",
		`{% if s == empty %}y{% else %}n{% endif %}`:       "y",
		`{% if a == empty %}y{% else %}n{% endif %}`:       "y",
		`{% if empty == nothing %}y{% else %}n{% endif %}`: "n",
	}
	for src, want := range cases {
		assertRender(t, src, want, data)
	}
}

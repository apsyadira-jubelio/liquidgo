package parity

import "testing"

// Ruby: Assign::Syntax = /(#{VariableSignature}+)\s*=\s*(.*)\s*/om — the m
// flag lets the value span lines (Dawn: `{%- assign x = y\n | filter: z -%}`).
func TestAssignValueSpansLines(t *testing.T) {
	assertRender(t, "{%- assign x = 'Ab'\n  | downcase\n  | append: 'c' -%}{{ x }}", "abc", nil)
}

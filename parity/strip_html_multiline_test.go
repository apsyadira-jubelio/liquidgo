package parity

import "testing"

// Ruby's STRIP_HTML_TAGS is /<.*?>/m, so a tag spanning lines is stripped.
func TestStripHTMLStripsMultilineTags(t *testing.T) {
	assertRender(t, `{{ s | strip_html }}`, "hi", map[string]interface{}{"s": "<a\nhref='x'>hi</a>"})
}

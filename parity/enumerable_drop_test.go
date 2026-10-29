package parity

import "testing"

// eachDrop is an Enumerable drop: not a slice, iterable only through Each
// (Shopify's global `collections`, `form.errors`).
type eachDrop struct{ items []interface{} }

func (d *eachDrop) Each(fn func(interface{})) {
	for _, it := range d.items {
		fn(it)
	}
}

// Ruby's InputIterator iterates any Enumerable (`input.is_a?(Enumerable)`),
// and VariableLookup calls first/size on anything that responds to them.
// liquidgo wrapped a non-slice drop as a single item, so
// `collections | sort: 'title'` saw one element.
func TestArrayFiltersIterateEnumerableDrops(t *testing.T) {
	data := func() map[string]interface{} {
		return map[string]interface{}{"c": &eachDrop{items: []interface{}{
			map[string]interface{}{"title": "b", "n": 2},
			map[string]interface{}{"title": "a", "n": 1},
			map[string]interface{}{"title": "c", "n": 3},
		}}, "extra": []interface{}{"d"}}
	}
	files["enum_item"] = "{{ x.title }}"
	cases := map[string]string{
		`{{ c | sort: 'title' | map: 'title' | join: ',' }}`:         "a,b,c",
		`{{ c | sort_natural: 'title' | map: 'title' | join: ',' }}`: "a,b,c",
		`{{ c | where: 'n', 2 | map: 'title' | join: ',' }}`:         "b",
		`{{ c | reject: 'n', 2 | map: 'title' | join: ',' }}`:        "a,c",
		`{{ c | reverse | map: 'title' | join: ',' }}`:               "c,a,b",
		`{{ c | uniq | size }}`:                                      "3",
		`{{ c | compact: 'title' | size }}`:                          "3",
		`{{ c | concat: extra | size }}`:                             "4",
		`{{ c | sum: 'n' }}`:                                         "6",
		`{{ c | map: 'title' | join: ',' }}`:                         "b,a,c",
		`{{ c | size }}`:                                             "3",
		`{% assign f = c | first %}{{ f.title }}`:                    "b",
		`{% assign l = c | last %}{{ l.title }}`:                     "c",
		`{{ c.size }}/{{ c.first.title }}/{{ c.last.title }}`:        "3/b/c",
		`{% for x in c %}{{ x.title }}{% endfor %}`:                  "bac",
		`{% render 'enum_item' for c as x %}`:                        "bac",
	}
	for src, want := range cases {
		assertRender(t, src, want, data())
	}
}

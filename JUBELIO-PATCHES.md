# Jubelio patches

Vendored from `github.com/TransistorFM/liquidgo` **v0.1.0** (commit `75d433e`;
module path `github.com/Notifuse/liquidgo`). Everything outside the patches
below is a verbatim copy of that tag. Each patch is independent so it can be
upstreamed as its own PR.

Pristine-tree test baseline (unchanged by these patches): `liquid`,
`liquid/tags` and `parity` have pre-existing failures (fork semantics vs old
upstream tests, `%Z` timezone) and `performance` needs the
`reference-liquid` submodule. `integration` and `liquid/tag` pass.

## 1. Render into `*strings.Builder` instead of `*string`

Every `RenderToOutputBuffer(context, output *string)` did `*output += s`, which
copies the whole output on each write: quadratic. A 5000-iteration `for`
(328 KB output) took ~300 ms and allocated 4.2 GB per render.

Change: the output parameter is `*strings.Builder` everywhere
(`RenderToOutputBuffer`, `Disabler`/`Disableable`, `renderNodeOptimized`, the
interface in `block_body.go`); `*output += s` became `output.WriteString(s)`.
Internal `Render()` helpers build into a local Builder. `RenderOptions.Output`
stays `*string` (public API unchanged there). Tests updated mechanically.

**API change:** custom tags must implement
`RenderToOutputBuffer(liquid.TagContext, *strings.Builder)`.

`Template.Render` also presizes the buffer to the previous render's length
(`outputHint`), so steady-state renders allocate the output about once.

`liquid/bigpage_bench_test.go` (Apple M4):

| | ns/op | B/op |
|---|---|---|
| before | 302,120,292 | 4,169,872,939 |
| after | ~750,000 | ~895,000 (output 327,786) |

## 2. Remove the global `VariableLookup` cache

`globalVariableLookupCache` was an unbounded process-wide `sync.Map` keyed by
markup: grows forever with theme/merchant-supplied templates. Parse already
memoises per template through the ParseContext expression cache, so the
global layer is dropped. `BenchmarkParseSection` (re-parsing the same source):
~885 µs / 615 KB before, ~885 µs / 617 KB after (+63 allocs of 13.6k).

Not changed: `globalExprCache` (expression.go) is still unbounded, but it is
only fed by `Context.Get` with a nil cache (constant strings in practice).

## 3. Lock `t.errors` in `Template.RenderToOutputBuffer` and `Errors()`

The deferred write of `t.errors` (and the lazy init in `Errors()`) ran without
`t.mu`, a data race when a cached template/partial renders concurrently.
Regression test: `liquid/jubelio_race_test.go` (fails under `-race` before).

## 4. Ruby parity fixes found by the storefront-go renderer port

No Ruby was available: expectations are reasoned from Liquid 5.x source
(and ActiveSupport for `blank?`, which Shopify loads), not run against the
gem. Each has a regression test in `parity/` that fails before the patch.

| Fix | Ruby semantics | Test |
|---|---|---|
| `==` panicked on two maps / two slices (`condition.go` `left == right`) | `==` is structural for Hash/Array (numeric elements across Integer/Float), identity for drops; never raises. Also used by `where`/`find`-style filters' `valuesEqual` | `parity/structural_equality_test.go` |
| `blank` was false for `[]`, `{}`, `false` | ActiveSupport `blank?`: nil, false, whitespace-only string, empty Array/Hash | `parity/blank_literal_test.go` |
| `FindVariable` checked environments before the outer scope, so a top-level `assign` could not shadow a global | `find_variable` checks all scopes, then environments; instance assigns are squashed with environments at `Context.new` (already ported). Upstream `TestCustomAssignsSquashInstanceAssigns` re-rendered the assigning template; it now re-parses `{{ foo }}` as Ruby's test does | `parity/assign_shadows_environment_test.go` |
| `obj[nil]` panicked on non-`map[string]` maps | `hash[nil]` is nil | `parity/nil_key_lookup_test.go` |
| Lax variable parse took everything before the first `\|` as the name (`a and b` → nil) | `MarkupWithQuotedFragment = /(QuotedFragment)(.*)/om`: name is the first fragment, the rest ignored unless it is a filter | `parity/lax_variable_first_fragment_test.go` |
| `at_least`/`at_most` returned Float for Integer input (`5000.0`) | `Utils.to_number` keeps Integer. Upstream `TestAtLeast`/`TestAtMost` updated to ints | `parity/at_least_at_most_integer_test.go` |
| `strip_html` tags regex lacked dotall | `STRIP_HTML_TAGS = /<.*?>/m` | `parity/strip_html_multiline_test.go` |

## 5. Cache per-call reflection on the render path

- `GetInvokableMethods`/`IsInvokable` (drop.go) rebuilt the method+field
  list via reflection on every lookup, including every map-key miss. Now a
  name set per `reflect.Type` in a `sync.Map` (bounded by distinct types).
  `GetInvokableMethods` returns a copy.
- `renderNodeOptimized` (block_body.go) did four `MethodByName` per tag
  render to detect `Render` overrides: cached per node type.
- Filter dispatch (strainer_template.go) did up to two `MethodByName` per
  filter instance per call: method index cached per (type, filter name);
  names reaching it are registered filters, so bounded.
- `strip_html`, `newline_to_br` and i18n interpolation compiled regexes per
  call: hoisted to package vars.
- `stringsTitle`/`snakeToCamel` allocate only when they change the string;
  string-keyed maps skip the reflective int-key path (a `ToInteger` error
  allocation per miss).

`liquid/jubelio_drop_bench_test.go` (Apple M4, 500 iterations per op):

| | ns/op | B/op | allocs/op |
|---|---|---|---|
| `BenchmarkDropLookups` before | 7,315,000 | 7,675,683 | 163,306 |
| `BenchmarkDropLookups` after | 800,000 | 303,171 | 16,298 |
| `BenchmarkFilterInvocation` before | 3,610,000 | 6,391,112 | 83,316 |
| `BenchmarkFilterInvocation` after | 1,450,000 | 618,126 | 35,804 |

## 6. Enumerable drops, multi-line assign, raw end tag, `nil == empty`, single block render

Same method as 4 (no Ruby; reasoned from Liquid 5.x source). Each test fails
before the patch. One upstream test (`TestConditionCheckMethodLiteral`,
`empty == nil`) updated to Ruby semantics.

| Fix | Ruby semantics | Test |
|---|---|---|
| Array filters wrapped a non-slice drop as one item (`collections \| sort: 'title'` saw one element). New `liquid.Eacher` (`Each(func(interface{}))`): `InputIterator` iterates it (sort, sort_natural, map, where, reject, reverse, uniq, compact, concat, sum, find…), `size`/`first`/`last`/`join` filters and `.size`/`.first`/`.last` lookups (unless the drop defines them) materialize it, `render … for` accepts it without `Count()`. `for` already did | `InputIterator#initialize`: `elsif input.is_a?(Enumerable) then input`; VariableLookup sends command methods the object responds to | `parity/enumerable_drop_test.go` |
| `assign` regex lacked `(?s)`: `{% assign x = a\n \| f %}` was a syntax error (Dawn) | `Assign::Syntax = /…(.*)\s*/om` | `parity/assign_multiline_test.go` |
| `raw` never closed on `{%- endraw -%}` (compared `"- endraw -"`), trimmed its body on `{%-`, and ignored `-%}` for the following text | `Raw::FullTokenPossiblyInvalid`; body verbatim; `trim_whitespace = token[-3] == '-'` (`raw_tag_test.rb` `test_output_in_raw`) | `parity/raw_whitespace_control_test.go` |
| `nil == empty` was true | `equal_variables`: `right.respond_to?(:empty?) ? right.empty? : nil`; nil/numbers never equal `empty` | `parity/nil_empty_test.go` |
| `Block.RenderToOutputBuffer` rendered the body three times (reflection on `*Block` always found `Block.Render`), so side effects in custom blocks (`increment`) ran repeatedly | Body renders once; `Render` overrides are dispatched by `renderNodeOptimized` | `parity/block_renders_once_test.go` |

## 7. Drops answer `empty?` / `blank?` themselves

Ruby `Condition` compares `x == empty` / `x == blank` by asking the value
(`respond_to?(:empty?)`). liquidgo only checked Go kinds, so a drop built on a
map (Shopify's `cart`, empty when it has no items) was never `empty` and
Dawn's empty-cart state never rendered. New `Emptier{Empty() bool}` and
`Blanker{Blank() bool}` interfaces are consulted first. Test:
`parity/empty_drop_test.go`.

## Known, not patched

- `Template.Render` merges top-level assigns back into the template
  (`instanceAssigns`) across renders — Ruby Liquid semantics; callers that
  cache templates should render via `Root()` with a fresh context.
- `empty`/`blank` comparisons don't call a drop's own `empty?`/`blank?`
  method (Ruby: `right.send(:empty?)` when it responds).

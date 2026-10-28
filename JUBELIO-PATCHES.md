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

## Known, not patched

- `Template.Render` merges top-level assigns back into the template
  (`instanceAssigns`) across renders — Ruby Liquid semantics; callers that
  cache templates should render via `Root()` with a fresh context.
- `Block.RenderToOutputBuffer` (block.go) renders the body up to three times
  via reflection to detect a `Render` override, for tags that do not override
  `RenderToOutputBuffer`. All built-in tags do.
- assign regex lacks `(?s)`; `raw` misses `{%- endraw -%}` (worked around in
  `internal/theme/engine`).

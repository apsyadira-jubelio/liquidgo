package liquid

import (
	"strings"
	"sync"
	"testing"
)

// A cached template rendered concurrently (e.g. a shared partial) must not
// race on t.errors. Run with -race.
func TestTemplateRenderToOutputBufferConcurrent(t *testing.T) {
	tmpl, err := ParseTemplate(`{{ x }}`, nil)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out strings.Builder
			tmpl.RenderToOutputBuffer(NewContext(), &out)
			_ = tmpl.Errors()
		}()
	}
	wg.Wait()
}

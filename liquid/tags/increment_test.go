package tags

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestIncrementTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewIncrementTag("increment", "counter", pc)
	if tag == nil {
		t.Fatal("Expected IncrementTag, got nil")
	}

	if tag.VariableName() != "counter" {
		t.Errorf("Expected variable name 'counter', got %q", tag.VariableName())
	}

	ctx := liquid.NewContext()
	var output strings.Builder

	// First call should output.String() 0
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "0" {
		t.Errorf("Expected '0', got %q", output.String())
	}

	// Second call should output.String() 1
	output.Reset()
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "1" {
		t.Errorf("Expected '1', got %q", output.String())
	}
}

// TestIncrementTagRenderToOutputBufferEdgeCases tests RenderToOutputBuffer edge cases
func TestIncrementTagRenderToOutputBufferEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewIncrementTag("increment", "counter", pc)

	ctx := liquid.NewContext()

	// Test with empty scopes (should initialize)
	ctx.Scopes() // Ensure scopes exist
	var output strings.Builder
	tag.RenderToOutputBuffer(ctx, &output)

	// Should output.String() 0 on first call
	if output.String() != "0" {
		t.Errorf("Expected '0' on first call, got %q", output.String())
	}

	// Test multiple increments
	for i := 1; i <= 5; i++ {
		output.Reset()
		tag.RenderToOutputBuffer(ctx, &output)
		expected := liquid.ToS(i, nil)
		if output.String() != expected {
			t.Errorf("Expected %q on call %d, got %q", expected, i+1, output.String())
		}
	}
}

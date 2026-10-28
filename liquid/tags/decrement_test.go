package tags

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestDecrementTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewDecrementTag("decrement", "counter", pc)
	if tag == nil {
		t.Fatal("Expected DecrementTag, got nil")
	}

	if tag.VariableName() != "counter" {
		t.Errorf("Expected variable name 'counter', got %q", tag.VariableName())
	}

	ctx := liquid.NewContext()
	var output strings.Builder

	// First call should output.String() -1
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "-1" {
		t.Errorf("Expected '-1', got %q", output.String())
	}

	// Second call should output.String() -2
	output.Reset()
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "-2" {
		t.Errorf("Expected '-2', got %q", output.String())
	}
}

// TestDecrementTagRenderToOutputBufferEdgeCases tests RenderToOutputBuffer edge cases
func TestDecrementTagRenderToOutputBufferEdgeCases(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag := NewDecrementTag("decrement", "counter", pc)

	ctx := liquid.NewContext()

	// Test with empty scopes (should initialize)
	ctx.Scopes() // Ensure scopes exist
	var output strings.Builder
	tag.RenderToOutputBuffer(ctx, &output)

	// Should output.String() -1 on first call
	if output.String() != "-1" {
		t.Errorf("Expected '-1' on first call, got %q", output.String())
	}

	// Test multiple decrements
	for i := 2; i <= 5; i++ {
		output.Reset()
		tag.RenderToOutputBuffer(ctx, &output)
		expected := liquid.ToS(-i, nil)
		if output.String() != expected {
			t.Errorf("Expected %q on call %d, got %q", expected, i, output.String())
		}
	}
}

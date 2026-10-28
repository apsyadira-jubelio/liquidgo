package tags

import (
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/liquid"
)

func TestIfchangedTag(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}
	if tag == nil {
		t.Fatal("Expected IfchangedTag, got nil")
	}
}

func TestIfchangedTagRendersWhenChanged(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}

	// Parse ifchanged block
	tokenizer := pc.NewTokenizer("content {% endifchanged %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output strings.Builder
	tag.RenderToOutputBuffer(ctx, &output)

	// First render should output.String() content
	if output.String() != "content " {
		t.Errorf("Expected output.String() 'content ', got %q", output.String())
	}
}

func TestIfchangedTagDoesNotRenderWhenUnchanged(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}

	// Parse ifchanged block
	tokenizer := pc.NewTokenizer("content {% endifchanged %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output strings.Builder

	// First render
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "content " {
		t.Errorf("Expected output.String() 'content ', got %q", output.String())
	}

	// Second render with same content should not output.String()
	output.Reset()
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "" {
		t.Errorf("Expected empty output.String() on second render, got %q", output.String())
	}
}

func TestIfchangedTagRendersWhenContentChanges(t *testing.T) {
	pc := liquid.NewParseContext(liquid.ParseContextOptions{})
	tag, err := NewIfchangedTag("ifchanged", "", pc)
	if err != nil {
		t.Fatalf("NewIfchangedTag() error = %v", err)
	}

	// Parse ifchanged block
	tokenizer := pc.NewTokenizer("{{item}} {% endifchanged %}", false, nil, false)
	err = tag.Parse(tokenizer)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	ctx := liquid.NewContext()
	var output strings.Builder

	// First render with item=1
	ctx.Set("item", 1)
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "1 " {
		t.Errorf("Expected output.String() '1 ', got %q", output.String())
	}

	// Second render with same item=1 should not output.String()
	output.Reset()
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "" {
		t.Errorf("Expected empty output.String(), got %q", output.String())
	}

	// Third render with item=2 should output.String()
	ctx.Set("item", 2)
	tag.RenderToOutputBuffer(ctx, &output)
	if output.String() != "2 " {
		t.Errorf("Expected output.String() '2 ', got %q", output.String())
	}
}

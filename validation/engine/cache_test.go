package engine

import "testing"

func TestCache_PutAndGet(t *testing.T) {
	c := NewCache()
	schema := map[string]any{"type": "string"}
	compiled, err := Compile(schema, CompileOpts{})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	c.Put("foo", compiled)
	if got, ok := c.Get("foo"); !ok || got != compiled {
		t.Fatalf("cached schema not returned: got=%v ok=%v", got, ok)
	}
	if _, ok := c.Get("bar"); ok {
		t.Fatal("Get for unknown key returned ok=true")
	}
}

func TestCache_GetOrCompile_UsesCache(t *testing.T) {
	c := NewCache()
	schema := map[string]any{"type": "integer"}
	first, err := c.GetOrCompile("k", schema, CompileOpts{})
	if err != nil {
		t.Fatalf("GetOrCompile: %v", err)
	}
	second, err := c.GetOrCompile("k", schema, CompileOpts{})
	if err != nil {
		t.Fatalf("GetOrCompile (cached): %v", err)
	}
	if first != second {
		t.Fatal("GetOrCompile did not return cached value on second call")
	}
	if c.Len() != 1 {
		t.Fatalf("Len = %d, want 1", c.Len())
	}
}

func TestCache_InvalidateAndReset(t *testing.T) {
	c := NewCache()
	schema := map[string]any{"type": "string"}
	for _, k := range []string{"a", "b", "c"} {
		if _, err := c.GetOrCompile(k, schema, CompileOpts{}); err != nil {
			t.Fatalf("GetOrCompile: %v", err)
		}
	}
	c.Invalidate("b")
	if _, ok := c.Get("b"); ok {
		t.Fatal("Invalidate did not remove entry")
	}
	if c.Len() != 2 {
		t.Fatalf("Len after Invalidate = %d, want 2", c.Len())
	}
	c.Reset()
	if c.Len() != 0 {
		t.Fatalf("Len after Reset = %d, want 0", c.Len())
	}
}

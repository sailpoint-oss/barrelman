package barrelman

import "testing"

type testPack struct {
	name   string
	called *int
}

func (p testPack) Name() string { return p.name }
func (p testPack) Register(reg *Registry) {
	if p.called != nil {
		*p.called = *p.called + 1
	}
}

func TestPluginRegistration(t *testing.T) {
	ClearPlugins()
	defer ClearPlugins()

	var a, b int
	RegisterPlugin(testPack{name: "pack-a", called: &a})
	RegisterPlugin(testPack{name: "pack-b", called: &b})

	if got := RegisteredPluginNames(); len(got) != 2 || got[0] != "pack-a" || got[1] != "pack-b" {
		t.Fatalf("RegisteredPluginNames = %v", got)
	}

	reg := NewRegistry()
	ApplyPlugins(reg)

	if a != 1 || b != 1 {
		t.Fatalf("expected each pack.Register to run once, got a=%d b=%d", a, b)
	}
}

func TestPluginReplaceByName(t *testing.T) {
	ClearPlugins()
	defer ClearPlugins()

	var first, second int
	RegisterPlugin(testPack{name: "pack", called: &first})
	RegisterPlugin(testPack{name: "pack", called: &second})

	reg := NewRegistry()
	ApplyPlugins(reg)
	if first != 0 || second != 1 {
		t.Fatalf("expected only second registration to run, got first=%d second=%d", first, second)
	}
}

func TestPluginNilSafe(t *testing.T) {
	ClearPlugins()
	defer ClearPlugins()
	RegisterPlugin(nil)
	if names := RegisteredPluginNames(); len(names) != 0 {
		t.Fatalf("nil plug-in must not register: %v", names)
	}
}

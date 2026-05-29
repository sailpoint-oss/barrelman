package barrelman

import "sync"

// RulePack is the contract a downstream consumer (Telescope, a private
// orchestration pipeline, an internal test harness) implements to inject
// extra rules into a Barrelman Registry without forking this package.
//
// Packs let the public Barrelman release ship only generic OAS / OWASP /
// formatting rules while leaving org-specific guideline checks, custom vendor
// extensions, and internal naming policies for downstream consumers to
// register privately at startup.
type RulePack interface {
	// Name returns a stable identifier for diagnostics and logging.
	Name() string

	// Register attaches every rule belonging to the pack to reg. The pack
	// MUST NOT depend on registration order with other packs.
	Register(reg *Registry)
}

var (
	pluginsMu sync.RWMutex
	plugins   []RulePack
)

// RegisterPlugin registers a RulePack to be applied by ApplyPlugins.
// Multiple registrations with the same name overwrite the previous entry,
// so a downstream pack can ship multiple versions during migration without
// causing duplicate-rule panics.
func RegisterPlugin(p RulePack) {
	if p == nil {
		return
	}
	pluginsMu.Lock()
	defer pluginsMu.Unlock()
	for i, existing := range plugins {
		if existing.Name() == p.Name() {
			plugins[i] = p
			return
		}
	}
	plugins = append(plugins, p)
}

// ApplyPlugins runs Register on every plug-in currently registered, in the
// order they were registered. Callers typically invoke this after
// constructing the in-process Registry and after analyzers.RegisterAll has
// installed the public rule set.
func ApplyPlugins(reg *Registry) {
	pluginsMu.RLock()
	defer pluginsMu.RUnlock()
	for _, p := range plugins {
		p.Register(reg)
	}
}

// RegisteredPluginNames returns the names of currently-registered plug-ins,
// in registration order. Useful for log lines and CLI diagnostics.
func RegisteredPluginNames() []string {
	pluginsMu.RLock()
	defer pluginsMu.RUnlock()
	out := make([]string, len(plugins))
	for i, p := range plugins {
		out[i] = p.Name()
	}
	return out
}

// ClearPlugins removes every registered plug-in. Intended for tests that
// want to exercise the registry without the global plug-in side-effects
// from blank imports. It is NOT safe to call concurrently with
// RegisterPlugin or ApplyPlugins.
func ClearPlugins() {
	pluginsMu.Lock()
	defer pluginsMu.Unlock()
	plugins = nil
}

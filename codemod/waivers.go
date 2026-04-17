package codemod

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Waiver represents a single entry in .telescope/waivers.yaml: the
// canonical rule slug, an optional JSON pointer narrowing the scope,
// the reason, an expiry timestamp, and the approving reviewer's
// handle. An empty Pointer waives every diagnostic emitted by Rule;
// a non-empty Pointer scopes the waiver to a specific element.
//
// Waivers are advisory for the codemod framework: when a fix would
// touch a waived target, the driver skips the patch. Diagnostics
// keep firing (so audit trails and waiver-expiry dashboards can
// surface them), but no automatic mutation occurs until the waiver
// lapses or is revoked.
type Waiver struct {
	Rule     string    `yaml:"rule"`
	Pointer  string    `yaml:"pointer,omitempty"`
	Reason   string    `yaml:"reason"`
	Until    time.Time `yaml:"until"`
	Approver string    `yaml:"approver,omitempty"`
}

// WaiverFile is the serialised form of .telescope/waivers.yaml.
type WaiverFile struct {
	Waivers []Waiver `yaml:"waivers"`
}

// LoadWaivers reads and decodes a waivers YAML file. Returns a
// zero-value (nil) WaiverSet when the file does not exist; other
// errors are returned verbatim so callers can decide whether to
// fail open or closed.
func LoadWaivers(path string) (*WaiverSet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &WaiverSet{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var wf WaiverFile
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	set := &WaiverSet{}
	for _, w := range wf.Waivers {
		set.waivers = append(set.waivers, w)
	}
	return set, nil
}

// DefaultWaiverPath returns the conventional path to the waivers file
// relative to root: `<root>/.telescope/waivers.yaml`. Callers that
// need a different location should resolve their own path.
func DefaultWaiverPath(root string) string {
	return filepath.Join(root, ".telescope", "waivers.yaml")
}

// WaiverSet is an indexed view over a list of Waiver entries.
// Zero value is safe to call Allows against (reports false for
// every input).
type WaiverSet struct {
	waivers []Waiver
	now     func() time.Time
}

// nowFn returns the configured clock or time.Now as a sensible default.
func (s *WaiverSet) nowFn() time.Time {
	if s == nil || s.now == nil {
		return time.Now()
	}
	return s.now()
}

// WithNow replaces the clock (test helper). Safe to call on a nil
// receiver because it returns a new set.
func (s *WaiverSet) WithNow(fn func() time.Time) *WaiverSet {
	if s == nil {
		return &WaiverSet{now: fn}
	}
	return &WaiverSet{waivers: s.waivers, now: fn}
}

// Allows reports whether the given (rule, pointer) pair is actively
// waived. An empty pointer on the waiver matches every pointer; a
// specific pointer matches exact equality only (no prefix semantics,
// to keep waivers explicit).
func (s *WaiverSet) Allows(rule, pointer string) bool {
	if s == nil {
		return false
	}
	now := s.nowFn()
	for _, w := range s.waivers {
		if w.Rule != rule {
			continue
		}
		if !w.Until.IsZero() && !now.Before(w.Until) {
			continue
		}
		if w.Pointer == "" || w.Pointer == pointer {
			return true
		}
	}
	return false
}

// Filter removes patches whose (RuleID, pointerFromPatch) match an
// active waiver. When the patch carries no pointer in Description
// (most Phase 2 patches don't), the rule-level waiver is the only
// applicable test. Returns a new slice; the input is not mutated.
func (s *WaiverSet) Filter(patches []Patch) []Patch {
	if s == nil || len(patches) == 0 {
		return patches
	}
	out := make([]Patch, 0, len(patches))
	for _, p := range patches {
		if s.Allows(p.RuleID, "") {
			continue
		}
		out = append(out, p)
	}
	return out
}

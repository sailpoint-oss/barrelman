package barrelman

import (
	"fmt"

	navigator "github.com/LukasParke/navigator"
)

// Reporter collects diagnostics during rule evaluation. Visitor callbacks
// receive a *Reporter and call At/AtRange to report issues.
type Reporter struct {
	id       string
	severity Severity
	diags    []Diagnostic

	pendingTags    []DiagnosticTag
	pendingRelated []RelatedInformation
	pendingData    interface{}
}

// NewReporter creates a Reporter for the given rule ID and default severity.
func NewReporter(id string, severity Severity) *Reporter {
	return &Reporter{id: id, severity: severity}
}

// WithTags sets diagnostic tags for the next reported diagnostic.
// Tags are consumed after the next At/AtRange/Error/Warn call.
func (r *Reporter) WithTags(tags ...DiagnosticTag) *Reporter {
	r.pendingTags = tags
	return r
}

// WithRelated adds related information for the next reported diagnostic.
func (r *Reporter) WithRelated(loc navigator.Loc, uri string, format string, args ...any) *Reporter {
	r.pendingRelated = append(r.pendingRelated, RelatedInformation{
		URI:     uri,
		Range:   loc.Range,
		Message: fmt.Sprintf(format, args...),
	})
	return r
}

// WithData sets an arbitrary data payload for the next reported diagnostic.
func (r *Reporter) WithData(data interface{}) *Reporter {
	r.pendingData = data
	return r
}

// At reports a diagnostic at the location of an OpenAPI model element.
func (r *Reporter) At(loc navigator.Loc, format string, args ...any) {
	r.report(loc.Range, r.severity, fmt.Sprintf(format, args...))
}

// AtRange reports a diagnostic at an explicit range.
func (r *Reporter) AtRange(rng Range, format string, args ...any) {
	r.report(rng, r.severity, fmt.Sprintf(format, args...))
}

// Error reports an error-severity diagnostic at the given location.
func (r *Reporter) Error(loc navigator.Loc, format string, args ...any) {
	r.report(loc.Range, SeverityError, fmt.Sprintf(format, args...))
}

// Warn reports a warning-severity diagnostic at the given location.
func (r *Reporter) Warn(loc navigator.Loc, format string, args ...any) {
	r.report(loc.Range, SeverityWarning, fmt.Sprintf(format, args...))
}

// ErrorAtRange reports an error-severity diagnostic at an explicit range.
func (r *Reporter) ErrorAtRange(rng Range, format string, args ...any) {
	r.report(rng, SeverityError, fmt.Sprintf(format, args...))
}

// WarnAtRange reports a warning-severity diagnostic at an explicit range.
func (r *Reporter) WarnAtRange(rng Range, format string, args ...any) {
	r.report(rng, SeverityWarning, fmt.Sprintf(format, args...))
}

// Diagnostics returns all reported diagnostics.
func (r *Reporter) Diagnostics() []Diagnostic {
	return r.diags
}

func (r *Reporter) report(rng Range, sev Severity, msg string) {
	if rng.End.Line > rng.Start.Line {
		rng.End = Position{Line: rng.Start.Line, Character: rng.Start.Character + 1000}
	}

	d := Diagnostic{
		Range:    rng,
		Severity: sev,
		Source:   Source,
		Code:     r.id,
		Message:  msg,
	}
	meta, ok := DefaultRegistry.Get(r.id)
	if ok && meta.DocURL != "" {
		d.CodeDescription = meta.DocURL
	}
	if len(r.pendingTags) > 0 {
		d.Tags = r.pendingTags
		r.pendingTags = nil
	}
	if len(r.pendingRelated) > 0 {
		d.Related = r.pendingRelated
		r.pendingRelated = nil
	}
	if r.pendingData != nil {
		d.Data = r.pendingData
		r.pendingData = nil
	}
	r.diags = append(r.diags, d)
}

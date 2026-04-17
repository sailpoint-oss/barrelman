// Package translate converts diagnostic types from adjacent validators
// (navigator, barrelman analyzers, spectral, vacuum) into the engine's
// unified Issue shape. It exists so consumers can aggregate issues from
// every pipeline stage into a single stream formatted with the engine's
// Zod-style message formatter.
package translate

import (
	navigator "github.com/sailpoint-oss/navigator"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/validation/engine"
)

// FromNavigator converts a navigator.Issue into an engine.Issue. Codes
// and messages are preserved; the RFC 6901 pointer is carried verbatim;
// range conversion keeps editor-facing (line, character) coordinates.
func FromNavigator(in navigator.Issue) engine.Issue {
	sev := engine.SeverityError
	switch in.Severity {
	case navigator.SeverityWarning:
		sev = engine.SeverityWarning
	case navigator.SeverityInfo:
		sev = engine.SeverityInfo
	}
	issue := engine.Issue{
		Code:     in.Code,
		Severity: sev,
		Source:   engine.SourceEngine,
		Message:  in.Message,
		Pointer:  in.Pointer,
	}
	issue.Path = issue.HumanPath()
	issue.Range = engine.Range{
		Start: engine.Position{Line: int(in.Range.Start.Line), Character: int(in.Range.Start.Character)},
		End:   engine.Position{Line: int(in.Range.End.Line), Character: int(in.Range.End.Character)},
	}
	return issue
}

// FromBarrelman converts a barrelman.Diagnostic into an engine.Issue.
// Barrelman diagnostics already carry range and code; this adapter
// renders the pointer (when present in Data) and the message as-is.
func FromBarrelman(in barrelman.Diagnostic) engine.Issue {
	sev := engine.SeverityError
	switch in.Severity {
	case barrelman.SeverityWarning:
		sev = engine.SeverityWarning
	case barrelman.SeverityInfo:
		sev = engine.SeverityInfo
	case barrelman.SeverityHint:
		sev = engine.SeverityHint
	}
	issue := engine.Issue{
		Code:     in.Code,
		Severity: sev,
		Source:   engine.Source(in.Source),
		Message:  in.Message,
	}
	if ptr, ok := issueDataString(in.Data, "pointer"); ok {
		issue.Pointer = ptr
		issue.Path = issue.HumanPath()
	}
	issue.Range = engine.Range{
		Start: engine.Position{Line: int(in.Range.Start.Line), Character: int(in.Range.Start.Character)},
		End:   engine.Position{Line: int(in.Range.End.Line), Character: int(in.Range.End.Character)},
	}
	return issue
}

// ToBarrelman converts an engine.Issue back into a barrelman.Diagnostic
// so downstream sinks that already consume barrelman.Diagnostic (LSP
// bridge, telescope report) can keep working unchanged.
func ToBarrelman(in engine.Issue) barrelman.Diagnostic {
	sev := barrelman.SeverityError
	switch in.Severity {
	case engine.SeverityWarning:
		sev = barrelman.SeverityWarning
	case engine.SeverityInfo:
		sev = barrelman.SeverityInfo
	case engine.SeverityHint:
		sev = barrelman.SeverityHint
	}
	data := map[string]any{}
	if in.Pointer != "" {
		data["pointer"] = in.Pointer
	}
	if in.Expected != "" {
		data["expected"] = in.Expected
	}
	if in.Received != "" {
		data["received"] = in.Received
	}
	if in.Suggestion != "" {
		data["suggestion"] = in.Suggestion
	}
	if len(data) == 0 {
		data = nil
	}
	return barrelman.Diagnostic{
		Code:     in.Code,
		Severity: sev,
		Source:   string(in.Source),
		Message:  in.Message,
		Range: barrelman.Range{
			Start: barrelman.Position{Line: uint32(in.Range.Start.Line), Character: uint32(in.Range.Start.Character)},
			End:   barrelman.Position{Line: uint32(in.Range.End.Line), Character: uint32(in.Range.End.Character)},
		},
		Data: data,
	}
}

func issueDataString(data any, key string) (string, bool) {
	switch d := data.(type) {
	case map[string]any:
		if v, ok := d[key]; ok {
			if s, ok := v.(string); ok {
				return s, true
			}
		}
	case map[string]string:
		if v, ok := d[key]; ok {
			return v, true
		}
	}
	return "", false
}

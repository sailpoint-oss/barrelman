package barrelman

import (
	"testing"

	navigator "github.com/sailpoint-oss/navigator"
)

func TestWalkNilIndex(t *testing.T) {
	r := NewReporter("test", SeverityWarning)
	Walk(nil, Visitors{
		Document: func(doc *navigator.Document, r *Reporter) {
			t.Error("Document visitor should not be called on nil index")
		},
	}, r)
	if len(r.Diagnostics()) != 0 {
		t.Error("expected no diagnostics for nil index")
	}
}

func TestWalkNilDocument(t *testing.T) {
	r := NewReporter("test", SeverityWarning)
	idx := &navigator.Index{}
	Walk(idx, Visitors{
		Document: func(doc *navigator.Document, r *Reporter) {
			t.Error("Document visitor should not be called on nil document")
		},
	}, r)
	if len(r.Diagnostics()) != 0 {
		t.Error("expected no diagnostics for nil document")
	}
}

func TestWalkDocumentVisitor(t *testing.T) {
	r := NewReporter("test", SeverityWarning)
	doc := &navigator.Document{}
	idx := &navigator.Index{Document: doc}

	called := false
	Walk(idx, Visitors{
		Document: func(d *navigator.Document, r *Reporter) {
			called = true
		},
	}, r)
	if !called {
		t.Error("Document visitor was not called")
	}
}

func TestWalkTagVisitor(t *testing.T) {
	r := NewReporter("test", SeverityWarning)
	doc := &navigator.Document{
		Tags: []navigator.Tag{
			{Name: "pets"},
			{Name: "users"},
		},
	}
	idx := &navigator.Index{Document: doc}

	var names []string
	Walk(idx, Visitors{
		Tag: func(tag *navigator.Tag, r *Reporter) {
			names = append(names, tag.Name)
		},
	}, r)
	if len(names) != 2 {
		t.Fatalf("expected 2 tag visits, got %d", len(names))
	}
	if names[0] != "pets" || names[1] != "users" {
		t.Errorf("tag names = %v, want [pets users]", names)
	}
}

func TestWalkServerVisitor(t *testing.T) {
	r := NewReporter("test", SeverityWarning)
	doc := &navigator.Document{
		Servers: []navigator.Server{
			{URL: "https://api.example.com"},
		},
	}
	idx := &navigator.Index{Document: doc}

	count := 0
	Walk(idx, Visitors{
		Server: func(srv *navigator.Server, r *Reporter) {
			count++
			if srv.URL != "https://api.example.com" {
				t.Errorf("unexpected server URL: %q", srv.URL)
			}
		},
	}, r)
	if count != 1 {
		t.Errorf("expected 1 server visit, got %d", count)
	}
}

func TestWalkCustomVisitor(t *testing.T) {
	r := NewReporter("test", SeverityWarning)
	doc := &navigator.Document{}
	idx := &navigator.Index{Document: doc}

	called := false
	Walk(idx, Visitors{
		Custom: func(i *navigator.Index, r *Reporter) {
			called = true
		},
	}, r)
	if !called {
		t.Error("Custom visitor was not called")
	}
}

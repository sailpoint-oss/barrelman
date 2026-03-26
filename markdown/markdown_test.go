package markdown

import "testing"

func TestValidateValidMarkdown(t *testing.T) {
	issues := Validate("# Title\n\nSome text.\n\n## Subtitle\n\nMore text.")
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateEmptyHeading(t *testing.T) {
	issues := Validate("# \n\nSome text.")
	found := false
	for _, iss := range issues {
		if iss.Message == "Empty heading" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'Empty heading' issue, got %v", issues)
	}
}

func TestValidateSkippedHeadingLevel(t *testing.T) {
	issues := Validate("# Title\n\n### Skipped h2")
	found := false
	for _, iss := range issues {
		if iss.Message == "Heading level skipped (expected h2, got h3)" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected heading level skipped issue, got %v", issues)
	}
}

func TestValidateEmptyLinkDestination(t *testing.T) {
	issues := Validate("Click [here]() for more.")
	found := false
	for _, iss := range issues {
		if iss.Message == "Link has empty destination" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected empty link destination issue, got %v", issues)
	}
}

func TestValidateEmptyImageSource(t *testing.T) {
	issues := Validate("![alt]()")
	found := false
	for _, iss := range issues {
		if iss.Message == "Image has empty source" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected empty image source issue, got %v", issues)
	}
}

func TestRender(t *testing.T) {
	html, err := Render("**bold** text")
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if html == "" {
		t.Error("Render returned empty string")
	}
}

func TestHeadings(t *testing.T) {
	headings := Headings("# Title\n\n## Subtitle\n\n### Section")
	if len(headings) != 3 {
		t.Fatalf("expected 3 headings, got %d", len(headings))
	}
	if headings[0].Level != 1 || headings[0].Text != "Title" {
		t.Errorf("heading[0] = %+v, want level=1, text=Title", headings[0])
	}
	if headings[1].Level != 2 || headings[1].Text != "Subtitle" {
		t.Errorf("heading[1] = %+v, want level=2, text=Subtitle", headings[1])
	}
}

func TestLinks(t *testing.T) {
	links := Links("Visit [example](https://example.com) and [docs](https://docs.example.com).")
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0].Destination != "https://example.com" {
		t.Errorf("link[0].Destination = %q", links[0].Destination)
	}
	if links[0].Text != "example" {
		t.Errorf("link[0].Text = %q", links[0].Text)
	}
}

func TestValidateNoIssues(t *testing.T) {
	issues := Validate("")
	if len(issues) != 0 {
		t.Errorf("empty string should have no issues, got %d", len(issues))
	}
}

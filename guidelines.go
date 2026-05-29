package barrelman

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	// DefaultGuidelinesBaseURL is the published generic guideline docs base URL.
	DefaultGuidelinesBaseURL = "https://telescope.dev/guidelines/"
	// GuidelinesBaseURLEnvVar lets local runs override the published site.
	GuidelinesBaseURLEnvVar = "TELESCOPE_GUIDELINES_BASE_URL"
)

var (
	guidelinesBaseURLMu sync.RWMutex
	guidelinesBaseURL   string
)

// SetGuidelinesBaseURL overrides the base URL used for external guideline links.
// Empty values clear the override and fall back to env/default resolution.
func SetGuidelinesBaseURL(raw string) {
	guidelinesBaseURLMu.Lock()
	defer guidelinesBaseURLMu.Unlock()
	guidelinesBaseURL = normalizeGuidelinesBaseURL(raw)
}

// GuidelinesBaseURL returns the effective external guideline docs base URL.
func GuidelinesBaseURL() string {
	guidelinesBaseURLMu.RLock()
	override := guidelinesBaseURL
	guidelinesBaseURLMu.RUnlock()
	if override != "" {
		return override
	}
	if env := normalizeGuidelinesBaseURL(os.Getenv(GuidelinesBaseURLEnvVar)); env != "" {
		return env
	}
	return DefaultGuidelinesBaseURL
}

// NormalizeGuidelineCode converts "104" or "rule-104" into "rule-104".
func NormalizeGuidelineCode(code string) string {
	id, ok := parseGuidelineID(code)
	if !ok {
		return ""
	}
	return fmt.Sprintf("rule-%03d", id)
}

// GuidelineIDFromCode extracts the numeric guideline id from "104" or "rule-104".
func GuidelineIDFromCode(code string) (int, bool) {
	return parseGuidelineID(code)
}

// GuidelineDocURL builds the published docs URL for a rule code.
func GuidelineDocURL(code string) string {
	id, ok := parseGuidelineID(code)
	if !ok {
		return ""
	}
	slug, ok := guidelineCategorySlug(id)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%sdocs/rules/%s#%d", GuidelinesBaseURL(), slug, id)
}

func normalizeGuidelinesBaseURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	return strings.TrimRight(trimmed, "/") + "/"
}

func parseGuidelineID(code string) (int, bool) {
	trimmed := strings.ToLower(strings.TrimSpace(code))
	trimmed = strings.TrimPrefix(trimmed, "#")
	trimmed = strings.TrimPrefix(trimmed, "rule-")
	if trimmed == "" {
		return 0, false
	}
	id, err := strconv.Atoi(trimmed)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func guidelineCategorySlug(id int) (string, bool) {
	switch {
	case id >= 100 && id <= 124:
		return "api-contract-and-documentation", true
	case id >= 200 && id <= 219:
		return "lifecycle-and-compatibility", true
	case id >= 300 && id <= 310:
		return "security-and-authorization", true
	case id >= 400 && id <= 418:
		return "http-semantics", true
	case id >= 500 && id <= 515:
		return "resource-modeling-and-urls", true
	case id >= 600 && id <= 606:
		return "requests-and-querying", true
	case id >= 700 && id <= 711:
		return "payload-conventions", true
	case id >= 800 && id <= 807:
		return "data-types-and-common-objects", true
	case id >= 900 && id <= 905:
		return "operations-and-quality", true
	default:
		return "", false
	}
}

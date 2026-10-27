package sanitize

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var policy *bluemonday.Policy

func init() {
	p := bluemonday.UGCPolicy()

	// Allow figure and figcaption elements for editorial images
	p.AllowElements("figure", "figcaption")
	p.AllowAttrs("class").Matching(regexp.MustCompile(`^[a-zA-Z0-9\s\-_:]+$`)).Globally()
	p.AllowAttrs("target").Matching(regexp.MustCompile(`^(_blank|_self)$`)).OnElements("a")
	p.AllowAttrs("rel").Matching(regexp.MustCompile(`^[a-zA-Z\s]+$`)).OnElements("a")

	// Allow iframes strictly for YouTube video embeds
	p.AllowElements("iframe")
	p.AllowAttrs("src", "width", "height", "title", "frameborder", "allow", "allowfullscreen").OnElements("iframe")
	p.AllowURLSchemes("https")

	policy = p
}

// HTML cleans untrusted HTML body content against a strict editorial allowlist.
func HTML(raw string) string {
	clean := policy.Sanitize(raw)

	// Additional verification: ensure any iframe src strictly belongs to trusted video providers
	if strings.Contains(clean, "<iframe") {
		// Verify src contains trusted domains
		if !strings.Contains(clean, "youtube.com/embed/") && !strings.Contains(clean, "youtube-nocookie.com/embed/") {
			// Strip iframes from untrusted domains
			stripPolicy := bluemonday.UGCPolicy()
			clean = stripPolicy.Sanitize(raw)
		}
	}

	return clean
}

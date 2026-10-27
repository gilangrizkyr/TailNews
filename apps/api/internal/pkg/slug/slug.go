package slug

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)
	multipleDashRegex    = regexp.MustCompile(`-+`)
)

// Generate converts any title string into a clean, URL-safe slug.
func Generate(title string) string {
	// Normalize to NFD (separates characters and accents)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, title)

	// Lowercase
	str := strings.ToLower(normalized)

	// Replace non-alphanumeric with dashes
	str = nonAlphanumericRegex.ReplaceAllString(str, "-")

	// Collapse multiple dashes
	str = multipleDashRegex.ReplaceAllString(str, "-")

	// Trim leading and trailing dashes
	str = strings.Trim(str, "-")

	// Limit length to 200 characters for safety
	if len(str) > 200 {
		str = str[:200]
		str = strings.TrimRight(str, "-")
	}

	if str == "" {
		return "artikel"
	}

	return str
}

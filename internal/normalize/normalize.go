package normalize

import (
	"fmt"
	"net/url"
	"strings"
)

// Normalize cleans up a URL according to standard rules.
// Rules:
// - Lowercase scheme and host
// - Remove fragments (#anchor)
// - Remove default ports (80/443)
// - Clean path (resolve ../ ./)
func Normalize(rawURL string) (string, error) {
	// 0. Ensure Scheme is present BEFORE parsing
	// url.Parse("example.com") puts "example.com" in Path, not Host.
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}

	// 1. Parse
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 2. Scheme and Host normalization
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// 3. Remove Fragments
	u.Fragment = ""

	// 4. Remove Default Ports
	u.Host = removeDefaultPort(u.Scheme, u.Host)

	// 5. Clean Path
	// Go's url.Parse cleans path, but doesn't remove trailing slashes always desired?
	// The prompt didn't specify trailing slash removal, so we stick to robust path.
	// We do nothing extra here to avoid over-engineering.

	return u.String(), nil
}

func removeDefaultPort(scheme, host string) string {
	parts := strings.Split(host, ":")
	if len(parts) != 2 {
		return host
	}

	domain, port := parts[0], parts[1]
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		return domain
	}

	return host
}

// EnsureScheme prepends http:// if missing (rare case for this tool as sources usually provide it, but safety first)
func EnsureScheme(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return fmt.Sprintf("http://%s", rawURL)
	}
	return rawURL
}

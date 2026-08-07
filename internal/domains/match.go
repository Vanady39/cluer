package domains

import (
	"net/url"
	"strings"
)

// MatchPath compares a target path against the current one.
//
// The pattern language is glob-like (`*` matches any run of characters), not a
// regular expression, because an admin writes it by hand in a form field. Given
// a regex field, they will eventually write a regex that is subtly wrong, and it
// will fail silently in production by matching nothing.
func MatchPath(pattern, path string) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return false
	}
	if pattern == "*" || pattern == "/*" {
		return true
	}

	// A pattern that says nothing about the query string is not asking about it.
	// Without this, "/dashboard" would fail to match "/dashboard?tab=stats" and
	// the admin would have no way to tell why their tour stopped appearing.
	if !strings.Contains(pattern, "?") {
		if i := strings.IndexByte(path, '?'); i >= 0 {
			path = path[:i]
		}
	}

	segments := strings.Split(pattern, "*")

	// No wildcard at all: exact match, ignoring a trailing slash so that
	// /dashboard and /dashboard/ are not treated as different pages.
	if len(segments) == 1 {
		return trimSlash(pattern) == trimSlash(path)
	}

	rest := path
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		switch {
		case i == 0:
			if !strings.HasPrefix(rest, seg) {
				return false
			}
			rest = rest[len(seg):]
		case i == len(segments)-1:
			return strings.HasSuffix(rest, seg)
		default:
			idx := strings.Index(rest, seg)
			if idx < 0 {
				return false
			}
			rest = rest[idx+len(seg):]
		}
	}
	return true
}

// PathFromURL extracts what the pattern is matched against: path plus query.
// The host is deliberately excluded — the same tour runs on localhost, staging
// and production, and pinning it to a hostname would break every promotion.
func PathFromURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	// No scheme and no host means the caller already handed us a path, so there
	// is nothing to strip. Checking for an empty Path instead would misread
	// "https://demo.local" — a URL whose path is the root — as unparseable.
	if parsed.Scheme == "" && parsed.Host == "" {
		if raw == "" {
			return "/"
		}
		return raw
	}

	path := parsed.Path
	if path == "" {
		path = "/"
	}
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	return path
}

func trimSlash(s string) string {
	if len(s) > 1 && strings.HasSuffix(s, "/") {
		return strings.TrimRight(s, "/")
	}
	return s
}

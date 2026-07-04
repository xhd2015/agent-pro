package ttywatch

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
)

const slugMaxBodyLen = 119
const slugHashSuffixLen = 9 // _ + 8 hex digits

// SlugifyCommandLine produces a stable slug from a session command argv.
func SlugifyCommandLine(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	parts := make([]string, 0, len(argv))
	for _, arg := range argv {
		token := strings.TrimLeft(arg, "-")
		parts = append(parts, sanitizeSlugToken(token))
	}
	slug := collapseSlugUnderscores(strings.Join(parts, "_"))
	slug = strings.Trim(slug, "_")
	if slug == "" {
		return ""
	}
	if len(slug) <= slugMaxBodyLen {
		return slug
	}
	hash := slugHash(slug)
	body := slug[:slugMaxBodyLen]
	body = strings.TrimRight(body, "_")
	return body + "_" + hash
}

// ServeSubcommand wraps SlugifyCommandLine as the internal reexec argv token.
func ServeSubcommand(argv []string) string {
	return "__serve_" + SlugifyCommandLine(argv) + "__"
}

// IsServeSubcommand reports whether arg is a slug-based serve reexec token.
func IsServeSubcommand(arg string) bool {
	return strings.HasPrefix(arg, "__serve_") && strings.HasSuffix(arg, "__") && arg != "__serve__"
}

func sanitizeSlugToken(token string) string {
	var b strings.Builder
	b.Grow(len(token))
	for _, r := range token {
		switch r {
		case '/', '\\':
			b.WriteByte('_')
		case ';', '|', '&', '$', '"', '\'', '<', '>', '(', ')', '{', '}', '[', ']', '*', '?', '~', '!', '#':
			b.WriteByte('_')
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '.' || r == '_' {
				b.WriteRune(r)
			} else {
				b.WriteByte('_')
			}
		}
	}
	return b.String()
}

func collapseSlugUnderscores(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	prevUnderscore := false
	for _, r := range s {
		if r == '_' {
			if prevUnderscore {
				continue
			}
			prevUnderscore = true
			b.WriteRune('_')
			continue
		}
		prevUnderscore = false
		b.WriteRune(r)
	}
	return b.String()
}

func slugHash(slug string) string {
	sum := sha256.Sum256([]byte(slug))
	return hex.EncodeToString(sum[:4])
}
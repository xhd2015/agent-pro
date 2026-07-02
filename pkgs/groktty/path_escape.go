package groktty

import "net/url"

func pathEscape(s string) string {
	return url.PathEscape(s)
}
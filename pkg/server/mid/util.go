package mid

import "strings"

var skipURLs = []string{
	"/metrics",
	"/health",
	"/static",
	"/proxy",
	"/favicon.ico",
}

func shouldSkipURL(path string) bool {
	for _, url := range skipURLs {
		if strings.HasPrefix(path, url) {
			return true
		}
	}
	return false
}

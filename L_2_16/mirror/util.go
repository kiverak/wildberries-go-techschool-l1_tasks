package mirror

import (
	"crypto/sha1"
	"encoding/hex"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func (c *Crawler) computeLocalPath(u *url.URL, contentType string) string {
	host := u.Hostname()
	p := u.EscapedPath()
	if p == "" || strings.HasSuffix(p, "/") {
		p = path.Join(p, "index.html")
	}
	ext := path.Ext(p)
	if ext == "" {
		if strings.Contains(contentType, "text/html") || strings.HasSuffix(p, "/") {
			p = p + "index.html"
		} else {
			if u.RawQuery != "" {
				h := sha1.Sum([]byte(u.RawQuery))
				p = p + "_" + hex.EncodeToString(h[:6])
			}
		}
	}
	if u.RawQuery != "" {
		h := sha1.Sum([]byte(u.RawQuery))
		p = p + "_" + hex.EncodeToString(h[:6])
	}
	p = strings.TrimPrefix(p, "/")
	return filepath.Join(host, filepath.FromSlash(p))
}

func sameHost(u1, u2 *url.URL) bool {
	return strings.EqualFold(u1.Hostname(), u2.Hostname())
}

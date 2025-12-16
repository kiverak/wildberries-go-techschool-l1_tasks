package mirror

import (
	"net/url"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

func (c *Crawler) processNode(n *html.Node, pageURL *url.URL, depth int) {
	if n.Type == html.ElementNode {
		for i, a := range n.Attr {
			isLink := false
			attrName := a.Key
			switch strings.ToLower(n.Data) {
			case "a":
				if attrName == "href" {
					isLink = true
				}
			case "img":
				if attrName == "src" {
					isLink = true
				}
			case "script":
				if attrName == "src" {
					isLink = true
				}
			case "link":
				if attrName == "href" {
					isLink = true
				}
			}
			if !isLink {
				continue
			}

			orig := a.Val
			if strings.HasPrefix(orig, "data:") || orig == "" {
				continue
			}
			resolved, err := pageURL.Parse(orig)
			if err != nil {
				continue
			}

			resolved.Fragment = ""

			if strings.ToLower(n.Data) == "a" {
				if sameHost(resolved, c.root) {
					c.mu.Lock()
					_, saved := c.savedPaths[resolved.String()]
					c.mu.Unlock()
					if !saved && depth < c.maxDepth {
						c.wg.Add(1)
						go func(u *url.URL, d int) {
							defer c.wg.Done()
							c.crawl(u, d+1)
						}(resolved, depth)
					}
					c.mu.Lock()
					local, ok := c.savedPaths[resolved.String()]
					c.mu.Unlock()
					if ok {
						rel, _ := filepath.Rel(filepath.Dir(filepath.Join(c.outDir, c.savedPaths[pageURL.String()])), filepath.Join(c.outDir, local))
						n.Attr[i].Val = filepath.ToSlash(rel)
					}
				}
			} else {
				if sameHost(resolved, c.root) {
					c.mu.Lock()
					_, saved := c.savedPaths[resolved.String()]
					c.mu.Unlock()
					if !saved {
						c.wg.Add(1)
						go func(u *url.URL) {
							defer c.wg.Done()
							c.crawl(u, depth)
						}(resolved)
					}
					c.mu.Lock()
					local, ok := c.savedPaths[resolved.String()]
					c.mu.Unlock()
					if ok {
						c.mu.Lock()
						srcLocal, ok2 := c.savedPaths[pageURL.String()]
						c.mu.Unlock()
						if ok2 {
							rel, _ := filepath.Rel(filepath.Dir(filepath.Join(c.outDir, srcLocal)), filepath.Join(c.outDir, local))
							n.Attr[i].Val = filepath.ToSlash(rel)
						}
					}
				}
			}
		}
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		c.processNode(child, pageURL, depth)
	}
}

func (c *Crawler) processNodeRecursive(n *html.Node, pageURL *url.URL, depth int) {
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			isLink := false
			attrName := a.Key
			switch strings.ToLower(n.Data) {
			case "a":
				if attrName == "href" {
					isLink = true
				}
			case "img":
				if attrName == "src" {
					isLink = true
				}
			case "script":
				if attrName == "src" {
					isLink = true
				}
			case "link":
				if attrName == "href" {
					isLink = true
				}
			}
			if !isLink {
				continue
			}

			orig := a.Val
			if strings.HasPrefix(orig, "data:") || orig == "" {
				continue
			}
			resolved, err := pageURL.Parse(orig)
			if err != nil {
				continue
			}

			resolved.Fragment = ""

			if strings.ToLower(n.Data) == "a" {
				if sameHost(resolved, c.root) {
					c.mu.Lock()
					_, saved := c.savedPaths[resolved.String()]
					c.mu.Unlock()
					if !saved && depth < c.maxDepth {
						c.wg.Add(1)
						go func(u *url.URL, d int) {
							defer c.wg.Done()
							c.crawl(u, d+1)
						}(resolved, depth)
					}
				}
			} else {
				if sameHost(resolved, c.root) {
					c.mu.Lock()
					_, saved := c.savedPaths[resolved.String()]
					c.mu.Unlock()
					if !saved {
						c.wg.Add(1)
						go func(u *url.URL) {
							defer c.wg.Done()
							c.crawl(u, depth)
						}(resolved)
					}
				}
			}
		}
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		c.processNodeRecursive(ch, pageURL, depth)
	}
}

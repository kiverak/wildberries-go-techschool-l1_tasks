package mirror

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"net/http"

	"golang.org/x/net/html"
)

type Crawler struct {
	root       *url.URL
	client     *http.Client
	outDir     string
	visited    map[string]struct{}
	savedPaths map[string]string // url -> local path
	mu         sync.Mutex
	sem        chan struct{}
	wg         sync.WaitGroup
	maxDepth   int
}

func NewCrawler(rootStr, outDir string, concurrency, maxDepth int) (*Crawler, error) {
	u, err := url.Parse(rootStr)
	if err != nil {
		return nil, err
	}
	c := &Crawler{
		root:       u,
		client:     &http.Client{Timeout: 15 * time.Second},
		outDir:     outDir,
		visited:    make(map[string]struct{}),
		savedPaths: make(map[string]string),
		sem:        make(chan struct{}, concurrency),
		maxDepth:   maxDepth,
	}
	return c, nil
}

func (c *Crawler) Acquire() { c.sem <- struct{}{} }
func (c *Crawler) Release() { <-c.sem }

func (c *Crawler) Start() error {
	if err := os.MkdirAll(c.outDir, 0755); err != nil {
		return err
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.crawl(c.root, 0)
	}()
	c.wg.Wait()
	return nil
}

func (c *Crawler) crawl(u *url.URL, depth int) {
	key := u.String()
	c.mu.Lock()
	if _, ok := c.visited[key]; ok {
		c.mu.Unlock()
		return
	}
	c.visited[key] = struct{}{}
	c.mu.Unlock()

	c.Acquire()
	resp, err := c.client.Get(u.String())
	c.Release()
	if err != nil {
		return
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	localPath := c.computeLocalPath(u, contentType)
	absLocalPath := filepath.Join(c.outDir, localPath)
	if err := os.MkdirAll(filepath.Dir(absLocalPath), 0755); err != nil {
		return
	}

	// If HTML, parse and rewrite links
	if strings.Contains(contentType, "text/html") || strings.HasSuffix(u.Path, "/") || filepath.Ext(u.Path) == "" {
		doc, err := html.Parse(strings.NewReader(string(body)))
		if err != nil {
			return
		}
		c.mu.Lock()
		c.savedPaths[key] = localPath
		c.mu.Unlock()

		c.processNode(doc, u, depth)

		// render back to bytes
		var b strings.Builder
		if err := html.Render(&b, doc); err != nil {
			return
		}
		if err := os.WriteFile(absLocalPath, []byte(b.String()), 0644); err != nil {
			return
		}
	} else {
		// resource: save as-is
		c.mu.Lock()
		c.savedPaths[key] = localPath
		c.mu.Unlock()
		if err := os.WriteFile(absLocalPath, body, 0644); err != nil {
			return
		}
	}
}

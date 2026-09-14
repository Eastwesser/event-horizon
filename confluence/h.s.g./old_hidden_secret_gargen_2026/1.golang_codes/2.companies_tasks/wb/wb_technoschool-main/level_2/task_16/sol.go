package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type job struct {
	u     *url.URL
	depth int
}

type config struct {
	root     *url.URL
	outDir   string
	maxDepth int
	parallel int
	client   *http.Client
}

var (
	reHref = regexp.MustCompile(`(?i)href\s*=\s*"([^"]+)"`)
	reSrc  = regexp.MustCompile(`(?i)src\s*=\s*"([^"]+)"`)
)

func sanitizePath(u *url.URL) string {
	p := u.Path
	if p == "" || strings.HasSuffix(p, "/") {
		p = filepath.Join(p, "index.html")
	}
	if strings.HasSuffix(p, "/") {
		p += "index.html"
	}
	if strings.HasSuffix(p, ".") {
		p += "html"
	}
	return filepath.Clean(p)
}

func ensureDir(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0o755)
}

func sameDomain(a, b *url.URL) bool {
	return strings.EqualFold(a.Hostname(), b.Hostname())
}

func resolve(base *url.URL, ref string) (*url.URL, error) {
	ru, err := url.Parse(ref)
	if err != nil {
		return nil, err
	}
	return base.ResolveReference(ru), nil
}

func writeFile(path string, r io.Reader) error {
	if err := ensureDir(path); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}

func rewriteLinks(html string, base *url.URL, cfg *config, discovered map[string]*url.URL) string {
	repl := func(m []string) string {
		raw := m[1]
		tu, err := resolve(base, raw)
		if err != nil {
			return m[0]
		}
		if tu.Scheme != cfg.root.Scheme || !sameDomain(cfg.root, tu) {
			return m[0]
		}
		local := sanitizePath(tu)
		discovered[tu.String()] = tu
		return strings.Replace(m[0], raw, local, 1)
	}
	html = reHref.ReplaceAllStringFunc(html, func(s string) string {
		sub := reHref.FindStringSubmatch(s)
		if len(sub) < 2 {
			return s
		}
		return repl(sub)
	})
	html = reSrc.ReplaceAllStringFunc(html, func(s string) string {
		sub := reSrc.FindStringSubmatch(s)
		if len(sub) < 2 {
			return s
		}
		return repl(sub)
	})
	return html
}

func fetch(u *url.URL, cfg *config) (*http.Response, error) {
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", "mini-wget/1.0")
	return cfg.client.Do(req)
}

func worker(id int, cfg *config, wg *sync.WaitGroup, q <-chan job, enqueue func(job), jobsWG *sync.WaitGroup) {
	defer wg.Done()
	for j := range q {
		func() {
			defer jobsWG.Done()
			if j.depth > cfg.maxDepth {
				return
			}
			resp, err := fetch(j.u, cfg)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			ct := resp.Header.Get("Content-Type")
			local := sanitizePath(j.u)
			full := filepath.Join(cfg.outDir, local)
			if strings.Contains(ct, "text/html") {
				b, _ := io.ReadAll(resp.Body)
				discovered := map[string]*url.URL{}
				rewritten := rewriteLinks(string(b), j.u, cfg, discovered)
				_ = writeFile(full, strings.NewReader(rewritten))
				if j.depth < cfg.maxDepth {
					keys := make([]string, 0, len(discovered))
					for k := range discovered {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						enqueue(job{u: discovered[k], depth: j.depth + 1})
					}
				}
			} else {
				_ = writeFile(full, resp.Body)
			}
		}()
	}
}

func main() {
	start := time.Now()
	depth := flag.Int("depth", 1, "max recursion depth")
	out := flag.String("out", "site", "output directory")
	par := flag.Int("parallel", 8, "parallel downloads")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: go run sol.go [flags] <url>")
		os.Exit(2)
	}
	rootStr := flag.Arg(0)
	root, err := url.Parse(rootStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if root.Scheme == "" {
		root.Scheme = "https"
	}
	cfg := &config{
		root:     root,
		outDir:   *out,
		maxDepth: *depth,
		parallel: *par,
		client:   &http.Client{Timeout: 15 * time.Second},
	}

	_ = os.MkdirAll(cfg.outDir, 0o755)

	mu := sync.Mutex{}
	visited := map[string]bool{}
	q := make(chan job, 1024)
	wg := sync.WaitGroup{}
	jobsWG := sync.WaitGroup{}

	enqueue := func(j job) {
		mu.Lock()
		if visited[j.u.String()] {
			mu.Unlock()
			return
		}
		visited[j.u.String()] = true
		mu.Unlock()
		jobsWG.Add(1)
		q <- j
	}

	for i := 0; i < cfg.parallel; i++ {
		wg.Add(1)
		go worker(i, cfg, &wg, q, enqueue, &jobsWG)
	}
	enqueue(job{u: root, depth: 0})
	go func() { jobsWG.Wait(); close(q) }()

	wg.Wait()
	fmt.Printf("done in %v, files: %d\n", time.Since(start), len(visited))
}

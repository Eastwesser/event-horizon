// BuildSiteMap: BFS по одному домену; внешние ссылки отбрасываем.
package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
)

// Мок «сайта» для офлайн-демо.
var pages = map[string]string{
	"https://w.ru/":        `<a href="https://w.ru/support"><a href="https://w.ru/news"><a href="https://other.com/x">`,
	"https://w.ru/news":    `<a href="https://w.ru/support"><a href="https://w.ru/">`,
	"https://w.ru/support": `<a href="https://w.ru/news">`,
}

func Get(_ context.Context, raw string) (string, error) {
	body, ok := pages[raw]
	if !ok {
		return "", fmt.Errorf("get %s: not found", raw)
	}
	return body, nil
}

func ParseHTML(body string) []string {
	var out []string
	for _, part := range strings.Split(body, `href="`) {
		if part == body {
			continue
		}
		end := strings.IndexByte(part, '"')
		if end > 0 {
			out = append(out, part[:end])
		}
	}
	return out
}

func ParseHostname(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

func BuildSiteMap(ctx context.Context, startURL string) map[string][]string {
	mainHost := ParseHostname(startURL)
	siteMap := make(map[string][]string)
	visited := map[string]bool{startURL: true}
	queue := []string{startURL}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]

		body, err := Get(ctx, cur)
		if err != nil {
			log.Println(err)
			siteMap[cur] = nil
			continue
		}

		var links []string
		seenLink := make(map[string]bool)
		for _, link := range ParseHTML(body) {
			if ParseHostname(link) != mainHost {
				continue
			}
			if seenLink[link] {
				continue
			}
			seenLink[link] = true
			links = append(links, link)
			if !visited[link] {
				visited[link] = true
				queue = append(queue, link)
			}
		}
		siteMap[cur] = links
	}
	return siteMap
}

func main() {
	m := BuildSiteMap(context.Background(), "https://w.ru/")
	for k, v := range m {
		fmt.Printf("%s → %v\n", k, v)
	}
}

package task11

import (
	"sort"
	"strings"
)

func FindAnagramSets(words []string) map[string][]string {
	type group struct {
		first string
		set   map[string]struct{}
	}

	sign := func(s string) string {
		r := []rune(s)
		sort.Slice(r, func(i, j int) bool { return r[i] < r[j] })
		return string(r)
	}

	g := make(map[string]*group)
	for _, w := range words {
		lw := strings.ToLower(w)
		key := sign(lw)
		if _, ok := g[key]; !ok {
			g[key] = &group{first: lw, set: make(map[string]struct{})}
		}
		g[key].set[lw] = struct{}{}
	}

	res := make(map[string][]string)
	for _, gr := range g {
		if len(gr.set) < 2 {
			continue
		}
		lst := make([]string, 0, len(gr.set))
		for w := range gr.set {
			lst = append(lst, w)
		}
		sort.Strings(lst)
		res[gr.first] = lst
	}
	return res
}

package rag

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Offline TF-IDF RAG over Prydwen markdown — no embedding API required.
// Good enough for "how does Outbox work in Shop?" style retrieval.

type Chunk struct {
	ID    string `json:"id"`
	Path  string `json:"path"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type Index struct {
	Chunks []Chunk             `json:"chunks"`
	DF     map[string]int      `json:"df"`
	TF     []map[string]float64 `json:"tf"` // per-chunk term frequencies
	N      int                 `json:"n"`
}

type Hit struct {
	Score float64 `json:"score"`
	Path  string  `json:"path"`
	Title string  `json:"title"`
	Text  string  `json:"text"`
}

var wordRe = regexp.MustCompile(`[[:alnum:]_]+`)

func tokenize(s string) []string {
	s = strings.ToLower(s)
	raw := wordRe.FindAllString(s, -1)
	out := make([]string, 0, len(raw))
	for _, w := range raw {
		if len(w) < 2 {
			continue
		}
		out = append(out, w)
	}
	return out
}

func BuildFromDir(root string) (*Index, error) {
	var chunks []Chunk
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		parts := splitChunks(string(b), 1200)
		for i, p := range parts {
			title := firstHeading(p)
			if title == "" {
				title = rel
			}
			chunks = append(chunks, Chunk{
				ID:    rel + "#" + itoa(i),
				Path:  rel,
				Title: title,
				Text:  strings.TrimSpace(p),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	idx := &Index{Chunks: chunks, DF: map[string]int{}, N: len(chunks)}
	idx.TF = make([]map[string]float64, len(chunks))
	for i, c := range chunks {
		tf := map[string]float64{}
		toks := tokenize(c.Title + " " + c.Text)
		for _, t := range toks {
			tf[t]++
		}
		seen := map[string]struct{}{}
		for t, n := range tf {
			tf[t] = n / float64(len(toks)+1)
			if _, ok := seen[t]; !ok {
				idx.DF[t]++
				seen[t] = struct{}{}
			}
		}
		idx.TF[i] = tf
	}
	return idx, nil
}

func (idx *Index) Search(query string, k int) []Hit {
	if idx == nil || idx.N == 0 {
		return nil
	}
	if k <= 0 {
		k = 5
	}
	qtf := map[string]float64{}
	qtoks := tokenize(query)
	for _, t := range qtoks {
		qtf[t]++
	}
	for t := range qtf {
		qtf[t] = qtf[t] / float64(len(qtoks)+1)
	}

	type scored struct {
		i int
		s float64
	}
	var scores []scored
	for i := range idx.Chunks {
		var dot, nq, nd float64
		for t, qv := range qtf {
			idf := math.Log(1 + float64(idx.N)/(1+float64(idx.DF[t])))
			qw := qv * idf
			dw := idx.TF[i][t] * idf
			dot += qw * dw
			nq += qw * qw
			nd += dw * dw
		}
		// also accumulate doc norm leftover lightly
		for t, dv := range idx.TF[i] {
			if _, ok := qtf[t]; ok {
				continue
			}
			idf := math.Log(1 + float64(idx.N)/(1+float64(idx.DF[t])))
			dw := dv * idf
			nd += dw * dw
		}
		if nq == 0 || nd == 0 {
			continue
		}
		cos := dot / (math.Sqrt(nq) * math.Sqrt(nd))
		if cos > 0 {
			scores = append(scores, scored{i: i, s: cos})
		}
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].s > scores[b].s })
	if len(scores) > k {
		scores = scores[:k]
	}
	hits := make([]Hit, 0, len(scores))
	for _, sc := range scores {
		c := idx.Chunks[sc.i]
		text := c.Text
		if len(text) > 800 {
			text = text[:800] + "…"
		}
		hits = append(hits, Hit{Score: sc.s, Path: c.Path, Title: c.Title, Text: text})
	}
	return hits
}

func (idx *Index) Save(path string) error {
	b, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func Load(path string) (*Index, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	return &idx, nil
}

func splitChunks(s string, maxRunes int) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	paras := strings.Split(s, "\n\n")
	var (
		out  []string
		cur  strings.Builder
		size int
	)
	flush := func() {
		t := strings.TrimSpace(cur.String())
		if t != "" {
			out = append(out, t)
		}
		cur.Reset()
		size = 0
	}
	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		r := utf8Len(p)
		if size > 0 && size+r > maxRunes {
			flush()
		}
		if size > 0 {
			cur.WriteString("\n\n")
		}
		cur.WriteString(p)
		size += r
		if size >= maxRunes {
			flush()
		}
	}
	flush()
	if len(out) == 0 && strings.TrimSpace(s) != "" {
		return []string{strings.TrimSpace(s)}
	}
	return out
}

func firstHeading(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	// first non-empty line
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return truncate(line, 80)
		}
	}
	return ""
}

func utf8Len(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	return string(b)
}

// EnsureIndex loads from cache or builds from root.
func EnsureIndex(root, cachePath string) (*Index, error) {
	if cachePath != "" {
		if idx, err := Load(cachePath); err == nil && idx.N > 0 {
			return idx, nil
		}
	}
	idx, err := BuildFromDir(root)
	if err != nil {
		return nil, err
	}
	if cachePath != "" {
		_ = idx.Save(cachePath)
	}
	return idx, nil
}

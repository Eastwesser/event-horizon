package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type options struct {
	after      int
	before     int
	context    int
	showNumber bool
	countOnly  bool
	ignoreCase bool
	invert     bool
	fixed      bool
}

func parseFlags() (options, map[string]bool) {
	opts := options{}
	seen := map[string]bool{}

	flag.IntVar(&opts.after, "A", 0, "print N lines of trailing context")
	flag.IntVar(&opts.before, "B", 0, "print N lines of leading context")
	flag.IntVar(&opts.context, "C", 0, "print N lines of output context")
	flag.BoolVar(&opts.countOnly, "c", false, "print only a count of matching lines")
	flag.BoolVar(&opts.ignoreCase, "i", false, "ignore case distinctions")
	flag.BoolVar(&opts.invert, "v", false, "invert the sense of matching")
	flag.BoolVar(&opts.fixed, "F", false, "interpret pattern as a fixed string")
	flag.BoolVar(&opts.showNumber, "n", false, "print line number with output lines")

	flag.Parse()
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })

	// Apply -C to both sides unless -A or -B explicitly provided
	if seen["C"] {
		if !seen["A"] {
			opts.after = opts.context
		}
		if !seen["B"] {
			opts.before = opts.context
		}
	}
	return opts, seen
}

func readAllLines(paths []string) ([]string, error) {
	if len(paths) == 0 {
		return readFrom(os.Stdin)
	}
	var out []string
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		lines, err := readFrom(f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		out = append(out, lines...)
	}
	return out, nil
}

func readFrom(r io.Reader) ([]string, error) {
	s := bufio.NewScanner(r)
	buf := make([]byte, 0, 1024*64)
	s.Buffer(buf, 1024*1024)
	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func buildMatcher(pattern string, opts options) (func(string) bool, error) {
	if opts.fixed {
		pat := pattern
		if opts.ignoreCase {
			pat = strings.ToLower(pat)
		}
		return func(line string) bool {
			if opts.ignoreCase {
				return strings.Contains(strings.ToLower(line), pat)
			}
			return strings.Contains(line, pat)
		}, nil
	}
	pat := pattern
	if opts.ignoreCase {
		pat = "(?i)" + pat
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}
	return func(line string) bool { return re.FindStringIndex(line) != nil }, nil
}

func unionRanges(ranges [][2]int, start, end int) [][2]int {
	if start > end {
		return ranges
	}
	if len(ranges) == 0 {
		return append(ranges, [2]int{start, end})
	}
	last := &ranges[len(ranges)-1]
	if start <= last[1]+1 { // overlap or touch
		if end > last[1] {
			last[1] = end
		}
		return ranges
	}
	return append(ranges, [2]int{start, end})
}

func run() error {
	opts, _ := parseFlags()
	args := flag.Args()
	if len(args) == 0 {
		return fmt.Errorf("pattern is required")
	}
	pattern := args[0]
	files := args[1:]

	lines, err := readAllLines(files)
	if err != nil {
		return err
	}

	matchFn, err := buildMatcher(pattern, opts)
	if err != nil {
		return err
	}

	matched := make([]bool, len(lines))
	for i, ln := range lines {
		ok := matchFn(ln)
		if opts.invert {
			ok = !ok
		}
		matched[i] = ok
	}

	if opts.countOnly {
		cnt := 0
		for _, m := range matched {
			if m {
				cnt++
			}
		}
		fmt.Println(cnt)
		return nil
	}

	var ranges [][2]int
	for i, m := range matched {
		if !m {
			continue
		}
		start := i - opts.before
		if start < 0 {
			start = 0
		}
		end := i + opts.after
		if end >= len(lines) {
			end = len(lines) - 1
		}
		ranges = unionRanges(ranges, start, end)
	}

	for _, rg := range ranges {
		for i := rg[0]; i <= rg[1]; i++ {
			if opts.showNumber {
				fmt.Printf("%d:%s\n", i+1, lines[i])
			} else {
				fmt.Println(lines[i])
			}
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

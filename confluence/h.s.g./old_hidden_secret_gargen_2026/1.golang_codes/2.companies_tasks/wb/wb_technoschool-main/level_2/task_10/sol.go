package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type options struct {
	column       int
	numeric      bool
	reverse      bool
	unique       bool
	month        bool
	ignoreTrail  bool
	checkSorted  bool
	humanNumeric bool
}

var monthToIndex = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

func parseFlags() options {
	opts := options{}
	flag.IntVar(&opts.column, "k", 0, "sort by column N (1-based), tab-separated")
	flag.BoolVar(&opts.numeric, "n", false, "compare according to numeric value")
	flag.BoolVar(&opts.reverse, "r", false, "reverse the result of comparisons")
	flag.BoolVar(&opts.unique, "u", false, "output only the first of an equal run")
	flag.BoolVar(&opts.month, "M", false, "compare month names")
	flag.BoolVar(&opts.ignoreTrail, "b", false, "ignore trailing blanks in comparisons")
	flag.BoolVar(&opts.checkSorted, "c", false, "check whether input is sorted; do not sort")
	flag.BoolVar(&opts.humanNumeric, "h", false, "compare human readable numbers (e.g. 2K, 3M)")
	flag.Parse()
	return opts
}

func readAllLines(paths []string) ([]string, error) {
	var lines []string
	if len(paths) == 0 {
		return readFrom(os.Stdin)
	}
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		ls, err := readFrom(f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
		lines = append(lines, ls...)
	}
	return lines, nil
}

func readFrom(r io.Reader) ([]string, error) {
	s := bufio.NewScanner(r)
	// Increase the buffer for long lines
	buf := make([]byte, 0, 1024*64)
	s.Buffer(buf, 1024*1024)
	var out []string
	for s.Scan() {
		out = append(out, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func extractKey(line string, column int, ignoreTrail bool) string {
	if column <= 0 {
		if ignoreTrail {
			return strings.TrimRightFunc(line, unicode.IsSpace)
		}
		return line
	}
	fields := strings.Split(line, "\t")
	var key string
	if column-1 < len(fields) {
		key = fields[column-1]
	}
	if ignoreTrail {
		key = strings.TrimRightFunc(key, unicode.IsSpace)
	}
	return key
}

func parseHumanNumber(s string) (float64, bool) {
	if s == "" {
		return 0, false
	}
	s = strings.TrimSpace(s)
	// Separate numeric part and suffix
	last := s[len(s)-1]
	mult := 1.0
	sfx := last
	snum := s
	if (last >= 'a' && last <= 'z') || (last >= 'A' && last <= 'Z') {
		snum = s[:len(s)-1]
		switch sfx {
		case 'k', 'K':
			mult = 1024
		case 'm', 'M':
			mult = 1024 * 1024
		case 'g', 'G':
			mult = 1024 * 1024 * 1024
		case 't', 'T':
			mult = 1024 * 1024 * 1024 * 1024
		default:
			// Unrecognized suffix; treat as plain number
			snum = s
		}
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(snum), 64)
	if err != nil {
		return 0, false
	}
	return v * mult, true
}

func parseNumber(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func parseMonth(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	key := strings.ToLower(strings.TrimSpace(s))
	if len(key) >= 3 {
		key = key[:3]
	}
	if idx, ok := monthToIndex[key]; ok {
		return idx, true
	}
	return 0, false
}

func lessFunc(lines []string, opts options) func(i, j int) bool {
	return func(i, j int) bool {
		ki := extractKey(lines[i], opts.column, opts.ignoreTrail)
		kj := extractKey(lines[j], opts.column, opts.ignoreTrail)

		// Decide comparison mode
		if opts.humanNumeric {
			vi, okI := parseHumanNumber(ki)
			vj, okJ := parseHumanNumber(kj)
			if okI && okJ {
				if opts.reverse {
					return vi > vj
				}
				return vi < vj
			}
		}
		if opts.numeric {
			vi, okI := parseNumber(ki)
			vj, okJ := parseNumber(kj)
			if okI && okJ {
				if opts.reverse {
					return vi > vj
				}
				return vi < vj
			}
		}
		if opts.month {
			mi, okI := parseMonth(ki)
			mj, okJ := parseMonth(kj)
			if okI && okJ {
				if opts.reverse {
					return mi > mj
				}
				return mi < mj
			}
		}

		// Fallback: lexicographic compare
		if opts.reverse {
			return kj < ki
		}
		return ki < kj
	}
}

func isSorted(lines []string, opts options) bool {
	lf := lessFunc(lines, opts)
	for i := 1; i < len(lines); i++ {
		if lf(i, i-1) { // current is less than previous -> not sorted
			return false
		}
	}
	return true
}

func uniqueInPlace(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	out := lines[:1]
	last := lines[0]
	for i := 1; i < len(lines); i++ {
		if lines[i] == last {
			continue
		}
		out = append(out, lines[i])
		last = lines[i]
	}
	return out
}

func run(opts options, args []string) error {
	lines, err := readAllLines(args)
	if err != nil {
		return err
	}
	if opts.checkSorted {
		if isSorted(lines, opts) {
			return nil
		}
		return errors.New("data is not sorted")
	}
	sort.SliceStable(lines, lessFunc(lines, opts))
	if opts.unique {
		lines = uniqueInPlace(lines)
	}
	for _, ln := range lines {
		fmt.Println(ln)
	}
	return nil
}

func main() {
	opts := parseFlags()
	if err := run(opts, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

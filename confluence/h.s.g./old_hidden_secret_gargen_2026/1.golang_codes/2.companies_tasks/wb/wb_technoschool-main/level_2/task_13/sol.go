package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type options struct {
	fieldsSpec string
	delim      string
	separated  bool
}

func parseFlags() options {
	opts := options{}
	flag.StringVar(&opts.fieldsSpec, "f", "", "fields to output, e.g. 1,3-5")
	flag.StringVar(&opts.delim, "d", "\t", "field delimiter (single rune)")
	flag.BoolVar(&opts.separated, "s", false, "only lines with delimiter")
	flag.Parse()
	return opts
}

func parseFields(spec string) map[int]struct{} {
	res := make(map[int]struct{})
	if spec == "" {
		return res
	}
	parts := strings.Split(spec, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(p, "-") {
			ab := strings.SplitN(p, "-", 2)
			if len(ab) != 2 {
				continue
			}
			a, errA := strconv.Atoi(strings.TrimSpace(ab[0]))
			b, errB := strconv.Atoi(strings.TrimSpace(ab[1]))
			if errA != nil || errB != nil || a <= 0 || b <= 0 {
				continue
			}
			if a > b {
				a, b = b, a
			}
			for i := a; i <= b; i++ {
				res[i] = struct{}{}
			}
			continue
		}
		n, err := strconv.Atoi(p)
		if err == nil && n > 0 {
			res[n] = struct{}{}
		}
	}
	return res
}

func cutLine(line string, opts options, wanted map[int]struct{}) (string, bool) {
	delim := []rune(opts.delim)
	if len(delim) == 0 {
		delim = []rune{'\t'}
	}
	sep := string(delim[0])
	if opts.separated && !strings.Contains(line, sep) {
		return "", false
	}
	fields := strings.Split(line, sep)
	var out []string
	for idx, f := range fields {
		if _, ok := wanted[idx+1]; ok {
			out = append(out, f)
		}
	}
	return strings.Join(out, sep), true
}

func main() {
	opts := parseFlags()
	if opts.fieldsSpec == "" {
		fmt.Fprintln(os.Stderr, "-f is required")
		os.Exit(1)
	}
	wanted := parseFields(opts.fieldsSpec)
	if len(wanted) == 0 {
		return
	}
	s := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*64)
	s.Buffer(buf, 1024*1024)
	for s.Scan() {
		line := s.Text()
		res, ok := cutLine(line, opts, wanted)
		if ok {
			fmt.Println(res)
		}
	}
	if err := s.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

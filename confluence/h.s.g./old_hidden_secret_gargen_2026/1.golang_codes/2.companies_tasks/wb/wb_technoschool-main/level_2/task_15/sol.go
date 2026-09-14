package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	runMu   sync.Mutex
	running []*exec.Cmd
)

func setRunning(cmds []*exec.Cmd) { runMu.Lock(); running = cmds; runMu.Unlock() }
func clearRunning()               { setRunning(nil) }

func interruptRunning() {
	runMu.Lock()
	cmds := append([]*exec.Cmd(nil), running...)
	runMu.Unlock()
	for _, c := range cmds {
		if c == nil || c.Process == nil {
			continue
		}
		_ = c.Process.Signal(os.Interrupt)
	}
}

func expandEnv(arg string) string {
	re := regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*`)
	return re.ReplaceAllStringFunc(arg, func(s string) string {
		name := s[1:]
		return os.Getenv(name)
	})
}

func splitOps(line string) []string {
	line = strings.ReplaceAll(line, "&&", " && ")
	line = strings.ReplaceAll(line, "||", " || ")
	line = strings.ReplaceAll(line, "|", " | ")
	line = strings.ReplaceAll(line, ">", " > ")
	line = strings.ReplaceAll(line, "<", " < ")
	fs := strings.Fields(line)
	for i := range fs {
		fs[i] = expandEnv(fs[i])
	}
	return fs
}

func isBuiltin(name string) bool {
	switch name {
	case "cd", "pwd", "echo", "kill", "ps":
		return true
	}
	return false
}

func runBuiltin(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if len(args) == 0 {
		return 0, nil
	}
	switch args[0] {
	case "cd":
		if len(args) < 2 {
			return 1, errors.New("cd: missing path")
		}
		p := args[1]
		if p == "~" {
			if h, ok := os.LookupEnv("HOME"); ok {
				p = h
			}
		}
		if !filepath.IsAbs(p) {
			cwd, _ := os.Getwd()
			p = filepath.Join(cwd, p)
		}
		if err := os.Chdir(p); err != nil {
			return 1, err
		}
		return 0, nil
	case "pwd":
		cwd, err := os.Getwd()
		if err != nil {
			return 1, err
		}
		fmt.Fprintln(stdout, cwd)
		return 0, nil
	case "echo":
		fmt.Fprintln(stdout, strings.Join(args[1:], " "))
		return 0, nil
	case "kill":
		if len(args) < 2 {
			return 1, errors.New("kill: missing pid")
		}
		pid, err := strconv.Atoi(args[1])
		if err != nil {
			return 1, err
		}
		p, err := os.FindProcess(pid)
		if err != nil {
			return 1, err
		}
		if err := p.Kill(); err != nil {
			return 1, err
		}
		return 0, nil
	case "ps":
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("tasklist")
		} else {
			cmd = exec.Command("ps", "-e", "-o", "pid,comm")
		}
		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return 1, err
		}
		return 0, nil
	}
	return 127, fmt.Errorf("unknown builtin: %s", args[0])
}

func parseStages(tokens []string) ([][]string, error) {
	var stages [][]string
	var cur []string
	for _, t := range tokens {
		if t == "|" {
			if len(cur) == 0 {
				return nil, errors.New("empty stage")
			}
			stages = append(stages, cur)
			cur = nil
			continue
		}
		cur = append(cur, t)
	}
	if len(cur) > 0 {
		stages = append(stages, cur)
	}
	return stages, nil
}

func applyRedirs(stage []string) (args []string, in io.Reader, out io.Writer, err error) {
	args = []string{}
	in = nil
	out = nil
	for i := 0; i < len(stage); i++ {
		t := stage[i]
		if t == ">" && i+1 < len(stage) {
			fname := stage[i+1]
			f, e := os.Create(fname)
			if e != nil {
				return nil, nil, nil, e
			}
			out = f
			i++
			continue
		}
		if t == "<" && i+1 < len(stage) {
			fname := stage[i+1]
			f, e := os.Open(fname)
			if e != nil {
				return nil, nil, nil, e
			}
			in = f
			i++
			continue
		}
		args = append(args, t)
	}
	return
}

func runPipeline(tokens []string) (int, error) {
	stages, err := parseStages(tokens)
	if err != nil {
		return 1, err
	}
	if len(stages) == 0 {
		return 0, nil
	}

	// Single stage builtins with redirections
	if len(stages) == 1 {
		args, inR, outW, err := applyRedirs(stages[0])
		if err != nil {
			return 1, err
		}
		if len(args) == 0 {
			return 0, nil
		}
		if isBuiltin(args[0]) {
			if inR == nil {
				inR = os.Stdin
			}
			if outW == nil {
				outW = os.Stdout
			}
			code, err := runBuiltin(args, inR, outW, os.Stderr)
			if c, ok := outW.(io.Closer); ok && c != os.Stdout {
				_ = c.Close()
			}
			return code, err
		}
	}

	var cmds []*exec.Cmd
	var prev io.Reader = os.Stdin
	var finalOut io.Writer = os.Stdout
	var closers []io.Closer

	for i, st := range stages {
		args, inR, outW, err := applyRedirs(st)
		if err != nil {
			return 1, err
		}
		if len(args) == 0 {
			return 1, errors.New("empty command")
		}
		name := args[0]
		argv := args[1:]
		cmd := exec.Command(name, argv...)
		if inR != nil {
			cmd.Stdin = inR
			if rc, ok := inR.(io.Closer); ok {
				closers = append(closers, rc)
			}
		} else {
			cmd.Stdin = prev
		}
		if i < len(stages)-1 {
			rc, err := cmd.StdoutPipe()
			if err != nil {
				return 1, err
			}
			prev = rc
		} else {
			if outW != nil {
				cmd.Stdout = outW
				if wc, ok := outW.(io.Closer); ok {
					closers = append(closers, wc)
				}
			} else {
				cmd.Stdout = finalOut
			}
		}
		cmd.Stderr = os.Stderr
		cmds = append(cmds, cmd)
	}

	setRunning(cmds)
	defer clearRunning()
	for _, c := range cmds {
		if err := c.Start(); err != nil {
			return 1, err
		}
	}
	var waitErr error
	for i := len(cmds) - 1; i >= 0; i-- {
		if err := cmds[i].Wait(); err != nil && waitErr == nil {
			waitErr = err
		}
	}
	for _, cl := range closers {
		_ = cl.Close()
	}
	if waitErr != nil {
		return 1, waitErr
	}
	return 0, nil
}

func runLine(line string) int {
	toks := splitOps(line)
	if len(toks) == 0 {
		return 0
	}
	var segments [][]string
	var ops []string
	cur := []string{}
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if t == "&&" || t == "||" {
			if len(cur) == 0 {
				return 1
			}
			segments = append(segments, cur)
			ops = append(ops, t)
			cur = nil
			continue
		}
		cur = append(cur, t)
	}
	if len(cur) > 0 {
		segments = append(segments, cur)
	}
	if len(segments) == 0 {
		return 0
	}

	prevCode := 0
	for i, seg := range segments {
		if i > 0 {
			sw := ops[i-1]
			if sw == "&&" && prevCode != 0 {
				continue
			}
			if sw == "||" && prevCode == 0 {
				continue
			}
		}
		code, _ := runPipeline(seg)
		prevCode = code
	}
	return prevCode
}

func main() {
	flag.Parse()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for range sig {
			interruptRunning()
		}
	}()

	s := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 0, 1024*64)
	s.Buffer(buf, 1024*1024)
	for {
		if !s.Scan() {
			if s.Err() != nil {
				fmt.Fprintln(os.Stderr, s.Err())
			}
			break
		}
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}
		_ = runLine(line)
	}
}

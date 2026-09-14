package main

import (
	"bufio"
	"fmt"
	"os"
)

func reverse(b []byte, i, j int) {
	for i < j {
		b[i], b[j] = b[j], b[i]
		i++
		j--
	}
}

func main() {
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	b := []byte(line)
	if len(b) == 0 {
		return
	}
	// normalize newlines to spaces
	for i := range b {
		if b[i] == '\n' || b[i] == '\r' {
			b[i] = ' '
		}
	}
	// trim leading/trailing spaces
	i, j := 0, len(b)-1
	for i <= j && b[i] == ' ' {
		i++
	}
	for j >= i && b[j] == ' ' {
		j--
	}
	if i > j {
		fmt.Println("")
		return
	}
	b = b[i : j+1]

	// compress multiple spaces to single
	w := 0
	prevSpace := false
	for _, ch := range b {
		if ch == ' ' {
			if !prevSpace {
				b[w] = ' '
				w++
				prevSpace = true
			}
		} else {
			b[w] = ch
			w++
			prevSpace = false
		}
	}
	b = b[:w]

	reverse(b, 0, len(b)-1)
	start := 0
	for i := 0; i <= len(b); i++ {
		if i == len(b) || b[i] == ' ' {
			reverse(b, start, i-1)
			start = i + 1
		}
	}
	fmt.Println(string(b))
}

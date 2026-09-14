// DLQ: после maxRetries сообщение уходит в dead-letter вместо бесконечного цикла.
package main

import "fmt"

type Msg struct {
	Body    string
	Attempt int
}

func process(m Msg) error {
	if m.Body == "poison" {
		return fmt.Errorf("cannot process")
	}
	return nil
}

func main() {
	const maxRetries = 3
	inbox := []Msg{{"ok", 0}, {"poison", 0}}
	var dlq []Msg

	for len(inbox) > 0 {
		m := inbox[0]
		inbox = inbox[1:]
		if err := process(m); err != nil {
			m.Attempt++
			if m.Attempt >= maxRetries {
				dlq = append(dlq, m)
				fmt.Println("→ DLQ:", m)
				continue
			}
			inbox = append(inbox, m) // retry
			fmt.Println("retry", m.Attempt, m.Body)
			continue
		}
		fmt.Println("acked:", m.Body)
	}
	fmt.Println("dlq size:", len(dlq))
}

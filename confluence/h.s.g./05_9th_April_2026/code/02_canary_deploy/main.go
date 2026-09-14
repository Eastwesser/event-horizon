// Canary: при росте error rate откатываем процент на предыдущую версию.
package main

import "fmt"

type Canary struct {
	percent   int // трафик на v2
	errRateV2 float64
}

func (c *Canary) Observe(errRate float64) {
	c.errRateV2 = errRate
	if errRate > 0.05 && c.percent > 0 {
		fmt.Printf("rollback: err=%.2f → percent %d→0\n", errRate, c.percent)
		c.percent = 0
		return
	}
	if errRate < 0.01 && c.percent < 100 {
		next := c.percent + 20
		if next > 100 {
			next = 100
		}
		fmt.Printf("promote: err=%.2f → percent %d→%d\n", errRate, c.percent, next)
		c.percent = next
	}
}

func main() {
	c := &Canary{}
	steps := []float64{0.002, 0.003, 0.004, 0.08, 0.001}
	for _, e := range steps {
		c.Observe(e)
		fmt.Println("  current canary %:", c.percent)
	}
}

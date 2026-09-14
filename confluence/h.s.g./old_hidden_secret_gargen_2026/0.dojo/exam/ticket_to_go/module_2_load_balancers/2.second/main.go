package main

import (
	"fmt"
	"hash/fnv"
	"sync"
)

/*
Другие алгоритмы LB (скажи trade-offs):

1) Weighted Round Robin — мощные ноды получают больше трафика.
2) Least Connections — куда меньше активных соединений (нужен счётчик).
3) IP Hash — sticky session без внешнего store (одинаковый IP → один бэкенд).
*/

// --- Weighted Round Robin (smooth, nginx-style упрощённо) ---

type weightedServer struct {
	name   string
	weight int
	cur    int
}

type WeightedRR struct {
	servers []*weightedServer
	total   int
	mu      sync.Mutex
}

func NewWeightedRR(weights map[string]int) *WeightedRR {
	w := &WeightedRR{}
	for name, wt := range weights {
		w.servers = append(w.servers, &weightedServer{name: name, weight: wt})
		w.total += wt
	}
	return w
}

func (w *WeightedRR) Next() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	var best *weightedServer
	for _, s := range w.servers {
		s.cur += s.weight
		if best == nil || s.cur > best.cur {
			best = s
		}
	}
	best.cur -= w.total
	return best.name
}

// --- Least Connections ---

type LeastConn struct {
	names []string
	conn  []int
	mu    sync.Mutex
}

func NewLeastConn(names ...string) *LeastConn {
	return &LeastConn{names: names, conn: make([]int, len(names))}
}

func (lc *LeastConn) Acquire() (name string, release func()) {
	lc.mu.Lock()
	best := 0
	for i := 1; i < len(lc.conn); i++ {
		if lc.conn[i] < lc.conn[best] {
			best = i
		}
	}
	lc.conn[best]++
	name = lc.names[best]
	idx := best
	lc.mu.Unlock()

	release = func() {
		lc.mu.Lock()
		lc.conn[idx]--
		lc.mu.Unlock()
	}
	return name, release
}

// --- IP Hash ---

func pickByIP(servers []string, ip string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(ip))
	return servers[int(h.Sum32())%len(servers)]
}

func main() {
	fmt.Println("--- Weighted RR (a:3, b:1) ---")
	wrr := NewWeightedRR(map[string]int{"a": 3, "b": 1})
	counts := map[string]int{}
	for i := 0; i < 8; i++ {
		n := wrr.Next()
		counts[n]++
		fmt.Println(n)
	}
	fmt.Println("counts:", counts)

	fmt.Println("--- Least Connections ---")
	lc := NewLeastConn("s1", "s2", "s3")
	var releases []func()
	for i := 0; i < 4; i++ {
		name, rel := lc.Acquire()
		fmt.Println("acquire", name)
		releases = append(releases, rel)
	}
	releases[0]()
	name, rel := lc.Acquire()
	fmt.Println("after free s?, got", name)
	rel()

	fmt.Println("--- IP Hash sticky ---")
	servers := []string{"eu-1", "eu-2", "us-1"}
	for _, ip := range []string{"1.1.1.1", "1.1.1.1", "8.8.8.8"} {
		fmt.Println(ip, "→", pickByIP(servers, ip))
	}
}

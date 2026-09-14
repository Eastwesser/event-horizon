package main
import ("fmt"; "sync/atomic")
type RR struct{ servers []string; i uint64 }
func (r *RR) Next() string {
	n := atomic.AddUint64(&r.i, 1) - 1
	return r.servers[n%uint64(len(r.servers))]
}
func main() {
	r := &RR{servers: []string{"a", "b", "c"}}
	for i := 0; i < 5; i++ { fmt.Println(r.Next()) }
}

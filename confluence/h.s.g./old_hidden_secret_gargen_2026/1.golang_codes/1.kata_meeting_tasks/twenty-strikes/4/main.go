package main
import "fmt"
// Strike 4: invert map[string]int → map[int][]string (dup values)
func invert(m map[string]int) map[int][]string {
	out := make(map[int][]string)
	for k, v := range m { out[v] = append(out[v], k) }
	return out
}
func main() {
	fmt.Println(invert(map[string]int{"a": 1, "b": 2, "c": 1}))
}

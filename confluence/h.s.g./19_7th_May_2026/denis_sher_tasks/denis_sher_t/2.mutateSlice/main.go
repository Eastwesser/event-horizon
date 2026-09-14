package main

import "fmt"

// mutate меняет underlying array через общий header (s[0]=100 видно снаружи),
// но append при нехватке cap аллоцирует новый массив — дальнейшие правки
// (s[1]=200) уже не видны снаружи. На собесе: «slice = ptr+len+cap, передаётся по значению».
func mutate(s []int) {
	s[0] = 100
	s = append(s, 4) // часто новый backing array, если cap==len
	s[1] = 200
}

func main() {
	data := []int{1, 2, 3} // len=3, cap=3 → append точно переаллоцирует
	mutate(data)
	fmt.Println(data) // [100 2 3]
}

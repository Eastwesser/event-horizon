// minCycle: длина кратчайшего простого цикла в неориентированном графе.
// На собесе: BFS от каждой вершины; при back-edge cycle = dist[u]+dist[v]+1.
package main

import "fmt"

func minCycle(n int, edges [][2]int) int {
	g := make([][]int, n+1)
	for _, e := range edges {
		a, b := e[0], e[1]
		g[a] = append(g[a], b)
		g[b] = append(g[b], a)
	}

	const inf = int(1e9)
	best := inf

	bfsFrom := func(start int) {
		dist := make([]int, n+1)
		parent := make([]int, n+1)
		for i := range dist {
			dist[i] = -1
			parent[i] = -1
		}
		q := []int{start}
		dist[start] = 0
		for len(q) > 0 {
			u := q[0]
			q = q[1:]
			for _, v := range g[u] {
				if dist[v] == -1 {
					dist[v] = dist[u] + 1
					parent[v] = u
					q = append(q, v)
					continue
				}
				if parent[u] != v {
					// back-edge → цикл
					if c := dist[u] + dist[v] + 1; c < best {
						best = c
					}
				}
			}
		}
	}

	for s := 1; s <= n; s++ {
		bfsFrom(s)
	}
	if best == inf {
		return -1
	}
	return best
}

func main() {
	// Пример 1 из tasks.md: треугольник → 3
	fmt.Println(minCycle(3, [][2]int{{1, 2}, {2, 3}, {1, 3}})) // 3

	// Без цикла: путь → -1
	fmt.Println(minCycle(4, [][2]int{{1, 2}, {2, 3}, {3, 4}})) // -1

	// Квадрат → 4
	fmt.Println(minCycle(4, [][2]int{{1, 2}, {2, 3}, {3, 4}, {4, 1}})) // 4

	// Треугольник + хвост → 3
	fmt.Println(minCycle(6, [][2]int{
		{1, 6}, {6, 4}, {4, 5}, {5, 2}, {2, 3}, {3, 1}, {1, 2},
	})) // 3
}

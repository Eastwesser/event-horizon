# UNIQ_RANDN

Нужна функция `uniqRandn(n int) []int`: вернуть слайс длины `n` из **уникальных** случайных int.

- `n <= 0` → пустой слайс.
- Диапазон генерации должен быть ≥ `n` (иначе бесконечный цикл).
- Типичный приём: `map[int]struct{}` как set + rejection sampling.

На собесе скажи: «set + rand, пока не набрали n».

**code:** [`../code/01_uniq_randn`](../code/01_uniq_randn)

См. также урок [`../../../../../00_4th_March_2026/code/01_uniq_randn`](../../../../../00_4th_March_2026/code/01_uniq_randn).

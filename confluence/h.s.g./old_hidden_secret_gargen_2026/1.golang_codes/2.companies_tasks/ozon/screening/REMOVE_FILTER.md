# REMOVE_FILTER

Обобщение `REMOVE_ZEROS`: отфильтровать слайс по предикату (например `v != 0` или `v > 0`) с compact in-place.

На собесе: «тот же write-index; предикат вынести в `func(int) bool`».

Если просят новый слайс — `append` в fresh; если in-place — не забывай про утечку хвоста (`s = s[:w]`).

**code:** [`../code/03_remove_zeros`](../code/03_remove_zeros) (частный случай predicate `!= 0`)

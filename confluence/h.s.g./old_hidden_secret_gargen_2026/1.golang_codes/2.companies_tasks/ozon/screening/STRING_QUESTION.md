# STRING_QUESTION

Screening про строки как immutable []byte / rune:

- `s[i]` — byte, не rune; для Unicode — `[]rune` или `for range`.
- Конкатенация в цикле → `strings.Builder`.
- `s[a:b]` шарит underlying bytes (как слайс).

Пример ловушки: правка через `[]rune("test")` и обратно в `string`.

Теория: [`../go/STRING.md`](../go/STRING.md). Runnable pack здесь нет.

# HFLabs — haven pointer

Код и разборы лежат в уроке 24 (не дублируем сюда).

→ [`../../../../24_unknown_date/`](../../../../24_unknown_date/)  
Файлы: `24_hflabs_task.md`, `24_tasks.md`, runnable в `code/`.

## Фокус на собесе

| Тема | Что говорить / где код |
|------|------------------------|
| **Java → Go review** | поля по регистру (export), нет Serializable → `encoding/json`+tags; `hashCode=1` убивает HashMap; в Go хеш ключа map сам по полям |
| **RLE** | A–Z only; count=1 без цифры; иначе `A4…`; ошибка на мусор → `code/08_rle` |
| **unique names** | первое уникальное `Name` → `code/09_unique_names` |
| **LRU** | «map + DLL, get двигает в head, O(1)» → `code/07_lru_cache` |

```bash
cd ../../../../24_unknown_date/code/08_rle
export GOWORK=off
go run main.go
```

См. также README урока: Java-разбор не переписывать — отвечать Go-фразами из таблицы выше.

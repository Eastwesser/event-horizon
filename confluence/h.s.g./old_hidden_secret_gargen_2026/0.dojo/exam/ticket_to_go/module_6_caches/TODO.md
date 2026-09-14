# Module 6 — Cache / LRU — статус haven

Путь: `0.dojo/exam/ticket_to_go/module_6_caches/`  
EH: Inventory Cache-Aside + Redis.

| # | Папка | Статус | Что сказать на собесе |
|---|--------|--------|------------------------|
| 1 | `1.task_warmup` | [x] | slice header; append может сменить backing array |
| 2 | `2.task_leetcode` | [x] | Trie: Insert/Search/StartsWith |
| 3 | `3.task_concurrency` | [x] | SafeLRU: Mutex (Get тоже двигает список) |
| 4a | `2.lru` | [x] | LRU = map + DLL, O(1) Get/Put |
| 4b | `1.first` | [x] | Write-Through vs Write-Back vs Cache-Aside |
| 5 | `5.task_sql` | [~] | PostGIS `<->` KNN / Haversine |

**Политики:** LRU / LFU / TTL.  
**Запись:** Through = consistency; Back = скорость + риск; Aside = контроль в приложении (EH).

Запуск: `export GOWORK=off && go run main.go`

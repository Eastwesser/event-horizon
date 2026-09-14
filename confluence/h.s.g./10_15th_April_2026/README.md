# Урок 10 — 15 апреля 2026

Sitemap crawler: карта сайта как `map[url][]links`, только свой домен, BFS/очередь + воркеры.

## Задачи

### 1. `BuildSiteMap`
Даны `Get` / `ParseHTML` / `ParseHostname`. Обход с `startURL`: скачать страницу, извлечь ссылки, внешние домены игнорировать, ошибки логировать и продолжать. Фразы на собесе:
- «visited set + очередь; объём маленький — всё в памяти»
- «воркер: Get → Parse → фильтр hostname → отдать edges»
- «ошибка Get ≠ паника: log + skip»
- «semaphore / лимит горутин — если сайт вырастет»

В демо `Get`/`ParseHTML` замоканы на фиксированный граф `w.ru`.

## Как запускать

```bash
export GOWORK=off
cd code/01_sitemap_crawler && go run main.go
```

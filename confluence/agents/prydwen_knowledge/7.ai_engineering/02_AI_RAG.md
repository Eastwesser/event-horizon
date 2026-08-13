# RAG — шпаргалка senior + Event Horizon Prydwen

## Идея

**Retrieval-Augmented Generation**: сначала найти релевантные фрагменты корпуса, потом дать их LLM как контекст. Модель отвечает опираясь на документы, а не только на веса обучения. Снижает галлюцинации по доменным фактам (порты, Outbox, subjects).

Пайплайн: ingest → chunk → index → retrieve(top-k) → prompt(context + question) → answer.

## Chunking

Режем документы так, чтобы кусок был самодостаточным и не раздувал контекст.

Практики:
- По заголовкам/секциям markdown (H1/H2), не «каждые N слов вслепую».
- Размер ~500–1500 символов/рун; overlap 10–15% если режете посередине абзаца.
- В EH Stage 2: ~1200-rune paragraphs/sections в `services/mcp/internal/rag`.
- Один топик на H1, конкретные имена сервисов (`Shop`, `Billing`), subjects (`shop.purchased`), порты — иначе retrieval слепой.

Плохой chunk: одна строка-заглушка или каша из пяти несвязанных тем.

## Retrieval

Запрос → вектор/лексический поиск → top-k чанков со score → (опционально) rerank.

Метрики качества: Recall@k (нашли ли нужный файл), MRR, ручной smoke «вопрос → ожидаемый path».

EH smoke: query «как работает Outbox в Shop» → hit `4.architecture_patterns/03_ARCH_INTEGRATION_PATTERNS.md`.

## TF-IDF vs embeddings

| | TF-IDF + cosine | Embeddings (dense) |
|--|-----------------|--------------------|
| Суть | Взвешивание редких терминов | Семантическая близость |
| Плюсы | Без vendor API, offline, дёшево, хорошо на jargon (`outbox`, `JetStream`) | Синонимы, перефраз, кросс-язык |
| Минусы | Слабее на «объясни своими словами» | Нужна модель/API, хранилище векторов, cost |
| EH | **Сейчас**: offline TF-IDF в `services/mcp/internal/rag` | Upgrade: OpenAI/Ollama + Qdrant/Chroma, тот же tool name |

Для внутренней инженерии с стабильным словарём (имена сервисов, паттерны) TF-IDF часто достаточно на Stage 2.

## Интерфейс в EH: `search_prydwen`

MCP tool возвращает top-k `{path, title, score, text}` по корпусу `confluence/agents/prydwen_knowledge`.

Проверка:
```bash
cd services/mcp && go test ./internal/rag/ -count=1
# через MCP: search_prydwen query="Outbox Shop"
```

Env: `PRYDWEN_ROOT`, опционально `RAG_INDEX_PATH` (кеш индекса).

## Авторские правила корпуса (чтобы RAG работал)

- Пишите факты, а не «см. код».
- Уникальные якоря: порты, subject names, имена таблиц.
- Не дублируйте противоречивые порты в разных файлах — один source of truth (`03_PROJECT_EVENT_HORIZON.md`).
- Избегайте one-line stubs: retrieval найдёт пустышку с высоким score по ключевому слову.

## Типичные ошибки RAG

- Слишком большие чанки → в top-k мало разнообразия.
- Нет фильтра по свежести → устаревший ORCHESTRATOR побеждает актуальный файл.
- LLM игнорирует контекст → усилить инструкцию «отвечай только по фрагментам, цитируй path».
- Индексация `.env` / секретов — запрещено.

## Upgrade path (рассказ на собесе)

1. Заменить TF-IDF векторы на embeddings.
2. Persist в Qdrant (`QDRANT_URL`), не меняя имя MCP tool.
3. Добавить OpenAPI / SQL schema dumps в корпус.
4. Hybrid search: BM25 + dense + reranker.

## Типичные вопросы на собесе

1. Зачем RAG, если есть fine-tuning?
2. Как выбрать размер chunk и overlap?
3. TF-IDF vs embeddings — когда что?
4. Как измерить качество retrieval до «красоты» ответа LLM?
5. Как устроен `search_prydwen` в Event Horizon?
6. Почему один топик на документ важен для RAG?
7. Что такое hybrid search?
8. Какие риски безопасности при индексации внутренней доки?

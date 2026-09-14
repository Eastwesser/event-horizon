# Модуль 1: Rate Limiter + Базовые алгоритмы
    - Разминка: Замыкания в горутинах (a, b, c вывод)
    - Литкод: Two Sum (LeetCode 1)
    - Конкурентность: Ограничение одновременных запросов (50 URL, ≤5 одновременно)
    - Паттерн: Token Bucket Rate Limiter (Allow() bool)
    - SQL: Поиск пользователей с >100 запросов/час
    - Проект: Rate limiting в Roolz/Adtime
 
# Модуль 2: Load Balancer + Структуры данных
    - Разминка: Работа с мапами (порядок итерации, конкурентность)
    - Литкод: Valid Parentheses (LeetCode 20)
    - Конкурентность: Worker Pool с graceful shutdown
    - Паттерн: Round Robin Load Balancer
    - SQL: Шардирование таблицы запросов по диапазонам
    - Проект: Как бы ты распределял нагрузку в Roolz (логистические расчёты)?
 
# Модуль 3: Circuit Breaker + Обработка ошибок
    - Разминка: Паника и recover
    - Литкод: Merge Two Sorted Lists (LeetCode 21)
    - Конкурентность: Graceful shutdown с контекстом
    - Паттерн: Circuit Breaker (Closed/Open/Half-Open)
    - SQL: Анализ ошибок запросов по времени
    - Проект: Circuit breaker в Roolz для внешних API поставщиков
 
# Модуль 4: Health Checker + Мониторинг
    - Разминка: Селекты на каналах с таймаутами
    - Литкод: Linked List Cycle (LeetCode 141)
    - Конкурентность: Периодические health checks нескольких сервисов
    - Паттерн: Health Checker с экспоненциальным backoff
    - SQL: SLA расчёты (uptime, error rate)
    - Проект: Мониторинг в Adtime (Prometheus, Grafana)
 
# Модуль 5: Retry with Backoff + Кэширование
    - Разминка: Работа с defer, порядок выполнения
    - Литкод: LRU Cache (LeetCode 146)
    - Конкурентность: Retry механизм с backoff для внешнего API
    - Паттерн: Экспоненциальный backoff + jitter
    - SQL: Кэширование результатов тяжёлых запросов
    - Проект: Retry в Adtime для генерации подарков
 
# Модуль 6: Кэширование (LRU) + Геолокация
    - Разминка: Срезы vs массивы, capacity
    - Литкод: Implement Trie (LeetCode 208) или Ball Tree задача
    - Конкурентность: Thread-safe LRU кэш
    - Паттерн: Write-through vs write-back кэш
    - SQL: Геолокационные запросы (ближайшие точки)
    - Проект: Ball Tree в Roolz для поиска поставщиков
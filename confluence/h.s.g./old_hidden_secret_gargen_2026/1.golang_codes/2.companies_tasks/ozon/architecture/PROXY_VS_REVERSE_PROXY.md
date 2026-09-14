## 4. `PROXY_VS_REVERSE_PROXY.md`

# Proxy vs Reverse Proxy

## Грейд
18

## Вопрос
В чем разница между proxy и reverse-proxy?

## Ответ

- **Proxy (forward proxy)** — клиент знает, что запрос/соединение пойдёт через proxy в точку назначения. 
Используется для обхода ограничений, анонимизации, кэширования.


- **Reverse proxy** — клиент ничего не знает о его существовании и не знает, кем конкретно будет обработан его запрос. 
Используется для load balancing, SSL termination, кэширования.

## Схема
Forward proxy:
Client → Proxy → Internet → Target Server
(Client знает про proxy)

Reverse proxy:
Client → Reverse Proxy → Internal Server 1
→ Internal Server 2
(Client думает, что общается напрямую с сервером)

## Что важно сказать на собеседовании

### Forward proxy (обычный прокси):
- Клиент **явно** настроен на его использование
- Применение:
  - Обход блокировок (firewall, geo-restrictions)
  - Анонимизация (скрывает реальный IP клиента)
  - Кэширование часто запрашиваемых ресурсов
  - Контроль доступа в корпоративной сети

### Reverse proxy (обратный прокси):
- Клиент **не знает** о его существовании
- Применение:
  - Load balancing (распределение нагрузки между серверами)
  - SSL termination (расшифровка HTTPS)
  - Кэширование ответов (уменьшение нагрузки на сервер)
  - Защита от DDoS (ограничение числа запросов)
  - Canary deployment (A/B тестирование)

## Примеры технологий

| Forward proxy | Reverse proxy |
|---------------|---------------|
| Squid | Nginx |
| Privoxy | HAProxy |
| SOCKS5 | Traefik |
| | Envoy (service mesh) |

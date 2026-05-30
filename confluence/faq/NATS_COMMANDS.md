# Терминал 1: NATS подписка (все события)
```bash
nats sub "score.updated" --server localhost:4222
```

# Терминал 2: Логи Game сервиса
```bash
tail -f /tmp/game.log | grep -E "Submitted|score|error|validated"
```

# Терминал 3: Логи Gateway (что приходит от фронтенда)
```bash
tail -f /tmp/gateway.log | grep -E "Gateway received|400|error"
```

# Терминал 4: Логи Billing (начисление наград)
```bash
tail -f /tmp/billing.log | grep -E "Added|tickets|lamps|error"
```

# Терминал 5: Логи Leaderboard (обновление топа)
```bash
tail -f /tmp/leaderboard.log | grep -E "Received|score|error"
```

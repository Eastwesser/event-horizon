# Сегодняшний план — шаг за шагом

Сделаем 3 маленьких шага, каждый с проверкой.

## Шаг 1: Починить userId в localStorage (15 минут)

✅ Открой frontend/src/components/Auth/Login.tsx и найди эту секцию:

```tsx
if (access_token) {
    localStorage.setItem('accessToken', access_token);
    localStorage.setItem('userId', user_id);  // 👈 должно быть раскомментировано
    // ...
}
```

## Что нужно сделать:

✅ - Убедись, что localStorage.setItem('userId', user_id) не закомментирован.
✅ - Добавь console.log('✅ Saved userId:', user_id); после сохранения.
✅ - Перезапусти фронтенд, очисти localStorage, залогинься.
✅ - Проверь в консоли: localStorage.getItem('userId')

ALL STEPS DONE

## Шаг 2: Починить суммирование очков в лидерборде (30 минут)

Проблема: сейчас лидерборд показывает отдельные рекорды, а не сумму.

Где править: services/leaderboard/internal/repository/redis_repo.go

Нужно изменить логику UpdateScore — вместо замены счёта, делать прибавку:

```go
func (r *RedisLeaderboardRepo) UpdateScore(ctx context.Context, gameID, userID, userEmail string, score int) (int, error) {
    key := fmt.Sprintf("leaderboard:%s", gameID)
    
    // 👇 Получаем текущий счёт
    currentScore, err := r.client.ZScore(ctx, key, userID).Result()
    if err == redis.Nil {
        currentScore = 0
    } else if err != nil {
        return 0, err
    }
    
    newScore := int(currentScore) + score  // 👈 СУММИРУЕМ, а не заменяем
    
    member := &redis.Z{
        Score:  float64(newScore),
        Member: userID,
    }
    
    if err := r.client.ZAdd(ctx, key, *member).Err(); err != nil {
        return 0, err
    }
    // ... остальное
}
```

## Шаг 3: WebSocket — постоянное соединение (20 минут)

В frontend/src/components/Leaderboard/Leaderboard.tsx:

Вынеси WebSocket из useEffect с пустым массивом зависимостей, чтобы он не пересоздавался:

```tsx
useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws/leaderboard');
    ws.onmessage = (event) => {
        // обновляем таблицу
    };
    return () => ws.close();
}, []); // 👈 пустой массив — соединение живёт всё время
```

А модальное окно открывается/закрывается независимо.

## Порядок работы

- Начинаем с Шага 1 (userId). Это самое простое и даст быстрый результат.
После проверки — коммитим.

- Затем Шаг 2 (суммирование очков).

- В конце Шаг 3 (WebSocket).

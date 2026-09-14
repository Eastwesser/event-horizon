# Паттерн: Concurrent Cache — Потокобезопасный кеш

## 🎯 Задача

Реализовать **потокобезопасный кеш** с возможностью одновременного чтения и записи из нескольких горутин.

## 🗡️ Проблема

```go
// ❌ НЕ ПОТОКОБЕЗОПАСНО
cache := make(map[string]interface{})

go func() { cache["key"] = "value" }()  // Write
go func() { value := cache["key"] }()   // Read
// → data race! Паника!
```

## 🔍 Решения

### 1. sync.RWMutex — Оптимизация для чтения

```go
type Cache struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]interface{}),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()         // Несколько горутин могут читать одновременно
	defer c.mu.RUnlock()
	
	val, ok := c.items[key]
	return val, ok
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()          // Только одна горутина может писать
	defer c.mu.Unlock()
	
	c.items[key] = value
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.items, key)
}
```

**Преимущества**:
- Множественное чтение без блокировки
- Простота реализации

**Недостатки**:
- Вся map блокируется при записи
- Нет TTL (time to live)

### 2. sync.Map — Встроенная concurrent map

```go
var cache sync.Map

// Set
cache.Store("key", "value")

// Get
value, ok := cache.Load("key")

// Delete
cache.Delete("key")

// Load or Store (атомарно)
actual, loaded := cache.LoadOrStore("key", "value")

// Range (итерация)
cache.Range(func(key, value interface{}) bool {
	fmt.Println(key, value)
	return true  // continue iteration
})
```

**Когда использовать**:
- Частое чтение, редкая запись
- Каждый ключ пишется один раз
- Несколько горутин читают/пишут **разные ключи**

**Не использовать**:
- Частая запись одних и тех же ключей
- Нужны сложные типы ключей (только `interface{}`)

### 3. Sharded Cache — Уменьшение contention

```go
type ShardedCache struct {
	shards []*Shard
	mask   uint32
}

type Shard struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

func NewShardedCache(numShards int) *ShardedCache {
	shards := make([]*Shard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = &Shard{
			items: make(map[string]interface{}),
		}
	}
	return &ShardedCache{
		shards: shards,
		mask:   uint32(numShards - 1),
	}
}

func (sc *ShardedCache) getShard(key string) *Shard {
	hash := fnv32(key)
	return sc.shards[hash&sc.mask]
}

func (sc *ShardedCache) Get(key string) (interface{}, bool) {
	shard := sc.getShard(key)
	shard.mu.RLock()
	defer shard.mu.RUnlock()
	
	val, ok := shard.items[key]
	return val, ok
}

func (sc *ShardedCache) Set(key string, value interface{}) {
	shard := sc.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()
	
	shard.items[key] = value
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= 16777619
		hash ^= uint32(key[i])
	}
	return hash
}
```

**Преимущества**:
- Уменьшение lock contention (16 shards → в 16 раз меньше блокировок)
- Масштабируется при высокой нагрузке

**Недостатки**:
- Сложнее код
- Больше памяти

### 4. Cache с TTL (Time To Live)

```go
type Item struct {
	Value      interface{}
	Expiration int64  // Unix timestamp
}

type TTLCache struct {
	mu    sync.RWMutex
	items map[string]Item
}

func NewTTLCache() *TTLCache {
	cache := &TTLCache{
		items: make(map[string]Item),
	}
	go cache.cleanupExpired()
	return cache
}

func (c *TTLCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	expiration := time.Now().Add(ttl).UnixNano()
	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *TTLCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}
	
	// Проверка TTL
	if time.Now().UnixNano() > item.Expiration {
		return nil, false
	}
	
	return item.Value, true
}

func (c *TTLCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now().UnixNano()
		for key, item := range c.items {
			if now > item.Expiration {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
```

## 💡 Правило Додзё

> **"Кеш — это река, в которую входят многие (читатели), но только один может изменить её русло (писатель). sync.RWMutex — это мост с двумя дорожками: широкая для чтения, узкая для записи."**

## 🧪 Производительность

```go
func BenchmarkCache(b *testing.B) {
	cache := NewCache()
	
	// Прогрев
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			if i%10 == 0 {
				cache.Set(key, i)  // 10% writes
			} else {
				cache.Get(key)     // 90% reads
			}
			i++
		}
	})
}
```

**Результаты** (на 8-core CPU):
```
sync.Mutex:      ~50M ops/s   (блокировка на каждое чтение)
sync.RWMutex:    ~200M ops/s  (параллельное чтение)
sync.Map:        ~150M ops/s  (оптимизирована для редкой записи)
ShardedCache:    ~300M ops/s  (меньше contention)
```

## 🎓 Реальные примеры

### 1. HTTP Cache Middleware

```go
type CacheMiddleware struct {
	cache *Cache
}

func (cm *CacheMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Path
		
		if cached, ok := cm.cache.Get(key); ok {
			w.Write(cached.([]byte))
			return
		}
		
		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)
		
		cm.cache.Set(key, rec.Body.Bytes())
		w.Write(rec.Body.Bytes())
	})
}
```

### 2. Memoization (кеширование результатов функций)

```go
type MemoizedFunc struct {
	cache *Cache
	fn    func(string) (interface{}, error)
}

func (mf *MemoizedFunc) Call(key string) (interface{}, error) {
	if cached, ok := mf.cache.Get(key); ok {
		return cached, nil
	}
	
	result, err := mf.fn(key)
	if err != nil {
		return nil, err
	}
	
	mf.cache.Set(key, result)
	return result, nil
}

// Использование
expensiveFunc := &MemoizedFunc{
	cache: NewCache(),
	fn: func(userID string) (interface{}, error) {
		return db.GetUser(userID)
	},
}

user, _ := expensiveFunc.Call("123")  // DB call
user, _ := expensiveFunc.Call("123")  // Cached
```

### 3. Singleflight (дедупликация запросов)

```go
import "golang.org/x/sync/singleflight"

type SingleflightCache struct {
	cache *Cache
	group singleflight.Group
}

func (sc *SingleflightCache) Get(key string, fetcher func() (interface{}, error)) (interface{}, error) {
	if cached, ok := sc.cache.Get(key); ok {
		return cached, nil
	}
	
	// Только первый вызов выполнит fetcher, остальные подождут
	result, err, _ := sc.group.Do(key, func() (interface{}, error) {
		value, err := fetcher()
		if err == nil {
			sc.cache.Set(key, value)
		}
		return value, err
	})
	
	return result, err
}
```

**Пример**:
```go
// 1000 горутин запрашивают один ключ одновременно
for i := 0; i < 1000; i++ {
	go func() {
		cache.Get("user:123", func() (interface{}, error) {
			return db.GetUser("123")  // Выполнится только 1 раз!
		})
	}()
}
```

## 🎓 Вопросы для медитации

1. Чем отличается `sync.Mutex` от `sync.RWMutex` в контексте кеша?
2. Почему `sync.Map` не подходит для случая "частая запись одного ключа"?
3. Как реализовать LRU (Least Recently Used) eviction policy?
4. Можно ли использовать channels вместо mutex для кеша?

## 🔗 Связанные паттерны

- **Lazy Initialization**: `sync.Once` для однократной инициализации
- **Read-Through Cache**: автоматическая загрузка из БД при промахе
- **Write-Through Cache**: запись в БД и кеш одновременно

## 📚 Best Practices

1. **RWMutex для read-heavy**: если чтений > 80%
2. **sync.Map для append-only**: если ключи пишутся один раз
3. **Sharded Cache для высокой нагрузки**: если > 10K ops/s
4. **TTL для временных данных**: сессии, токены, OTP
5. **Singleflight для внешних API**: избежать thundering herd

---

**Задание**: Реализуй Cache с поддержкой:
1. **LRU eviction**: удаление самого старого неиспользуемого элемента
2. **Max size**: ограничение количества элементов
3. **TTL per key**: каждый ключ со своим временем жизни
4. **Metrics**: hit rate, miss rate, eviction count

Объясни, почему нельзя использовать `map[string]interface{}` с `sync.Mutex` для всех случаев.

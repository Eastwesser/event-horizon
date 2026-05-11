package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "runtime"
    "sync"
    "sync/atomic"
    "syscall"
    "time"

    "github.com/gorilla/websocket"
)

var (
    successCount int64
    failCount    int64
    activeCount  int64
)

func main() {
    numPtr := flag.Int("n", 100, "number of connections")
    durationPtr := flag.Int("d", 30, "duration in seconds")
    verbosePtr := flag.Bool("v", false, "verbose output")
    flag.Parse()

    num := *numPtr
    duration := *durationPtr
    verbose := *verbosePtr

    fmt.Printf("🔌 Starting %d WebSocket connections for %d seconds\n", num, duration)
    fmt.Printf("📊 Gateway endpoint: ws://localhost:8080/ws/leaderboard\n\n")

    // Увеличиваем лимит (проверка)
    fmt.Printf("🔧 File descriptor limit: %d\n", getFDLimit())

    startTime := time.Now()

    // Канал для сигнала завершения
    done := make(chan bool)

    // Graceful shutdown
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    // Запускаем соединения
    var wg sync.WaitGroup
    for i := 0; i < num; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            connectWebSocket(id, duration, verbose)
        }(i)
        
        // Небольшая задержка между подключениями, чтобы не перегружать gateway
        if i%100 == 0 && i > 0 {
            time.Sleep(10 * time.Millisecond)
        }
    }

    // Мониторинг ресурсов
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        for range ticker.C {
            printStats(num, duration, startTime)
        }
    }()

    // Ждём завершения или сигнала
    go func() {
        wg.Wait()
        done <- true
    }()

    select {
    case <-sigCh:
        fmt.Println("\n⚠️ Interrupted by user")
    case <-done:
        fmt.Println("\n✅ Test completed")
    }

    // Финальная статистика
    printFinalStats(duration)
}

func connectWebSocket(id int, duration int, verbose bool) {
    url := "ws://localhost:8080/ws/leaderboard"
    conn, _, err := websocket.DefaultDialer.Dial(url, nil)
    if err != nil {
        atomic.AddInt64(&failCount, 1)
        if verbose {
            log.Printf("[%d] ❌ Failed to connect: %v", id, err)
        }
        return
    }
    defer conn.Close()

    atomic.AddInt64(&successCount, 1)
    atomic.AddInt64(&activeCount, 1)
    if verbose {
        log.Printf("[%d] ✅ Connected. Active: %d", id, atomic.LoadInt64(&activeCount))
    }

    // Читаем сообщения в фоне
    go func() {
        for {
            _, msg, err := conn.ReadMessage()
            if err != nil {
                atomic.AddInt64(&activeCount, -1)
                if verbose {
                    log.Printf("[%d] 🔴 Disconnected. Active: %d", id, atomic.LoadInt64(&activeCount))
                }
                return
            }
            if verbose && len(msg) > 0 {
                log.Printf("[%d] 📨 Received: %s", id, string(msg))
            }
        }
    }()

    time.Sleep(time.Duration(duration) * time.Second)
}

func printStats(total, duration int, startTime time.Time) {
    success := atomic.LoadInt64(&successCount)
    fail := atomic.LoadInt64(&failCount)
    active := atomic.LoadInt64(&activeCount)
    elapsed := time.Since(startTime)

    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)

    fmt.Printf("\n📊 Stats at %.0fs: ", elapsed.Seconds())
    fmt.Printf("Connected: %d/%d, ", success, total)
    fmt.Printf("Active: %d, ", active)
    fmt.Printf("Failed: %d, ", fail)
    fmt.Printf("Memory: %.1fMB", float64(mem.Alloc)/1024/1024)
    fmt.Println()
}

func printFinalStats(duration int) {
    success := atomic.LoadInt64(&successCount)
    fail := atomic.LoadInt64(&failCount)
    active := atomic.LoadInt64(&activeCount)

    var mem runtime.MemStats
    runtime.ReadMemStats(&mem)

    fmt.Println("\n" + stringRepeat("=", 50))
    fmt.Println("📊 FINAL RESULTS")
    fmt.Println(stringRepeat("=", 50))
    fmt.Printf("✅ Successful connections: %d\n", success)
    fmt.Printf("❌ Failed connections: %d\n", fail)
    fmt.Printf("🔌 Peak active connections: %d\n", active)
    fmt.Printf("💾 Memory usage: %.1fMB\n", float64(mem.Alloc)/1024/1024)
    fmt.Printf("🔧 File descriptor limit: %d\n", getFDLimit())
    fmt.Println(stringRepeat("=", 50))
}

func getFDLimit() int {
    var rlim syscall.Rlimit
    err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
    if err != nil {
        return 0
    }
    return int(rlim.Cur)
}

func stringRepeat(s string, count int) string {
    result := ""
    for i := 0; i < count; i++ {
        result += s
    }
    return result
}
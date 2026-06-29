package balancer

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "sync"
    "sync/atomic"

    "github.com/Eastwesser/event-horizon/services/balancer/internal/metrics"
)

type Backend struct {
    URL         *url.URL
    ActiveConns int32
    Mutex       sync.Mutex
    Proxy       *httputil.ReverseProxy
}

type LeastConnBalancer struct {
    backends []*Backend
    mu       sync.RWMutex
}

func NewLeastConnBalancer(urls []string) *LeastConnBalancer {
    backends := make([]*Backend, 0, len(urls))

    for _, rawURL := range urls {
        u, err := url.Parse(rawURL)
        if err != nil {
            log.Printf("Failed to parse URL %s: %v", rawURL, err)
            continue
        }

        backend := &Backend{
            URL:   u,
            Proxy: httputil.NewSingleHostReverseProxy(u),
        }

        // custom proxy for logging
        backend.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
            log.Printf("Proxy error: %v", err)
            http.Error(w, "Backend error", http.StatusBadGateway)
        }

        backends = append(backends, backend)
    }

    return &LeastConnBalancer{
        backends: backends,
    }
}

func (lb *LeastConnBalancer) getLeastConnBackend() *Backend {
    lb.mu.RLock()
    defer lb.mu.RUnlock()

    if len(lb.backends) == 0 {
        return nil
    }

    var selected *Backend
    var minConns int32 = 2147483647 // MaxInt32

    for _, b := range lb.backends {
        conns := atomic.LoadInt32(&b.ActiveConns)
        if conns < minConns {
            minConns = conns
            selected = b
        }
    }

    return selected
}

func (lb *LeastConnBalancer) getTotalActiveConns() int32 {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    var total int32
    for _, b := range lb.backends {
        total += atomic.LoadInt32(&b.ActiveConns)
    }
    return total
}

func (lb *LeastConnBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    metrics.RequestsTotal.Inc()
    metrics.ActiveConnections.Set(float64(lb.getTotalActiveConns()))
    backend := lb.getLeastConnBackend()

    if backend == nil {
        http.Error(w, "No available backends", http.StatusServiceUnavailable)
        return
    }

    atomic.AddInt32(&backend.ActiveConns, 1)
    defer atomic.AddInt32(&backend.ActiveConns, -1)

    log.Printf("🔀 Proxying to %s (active: %d)", backend.URL.Host, backend.ActiveConns)

    // Добавляем заголовок с именем бэкенда
    r.Header.Set("X-Balancer-Backend", backend.URL.Host)

    backend.Proxy.ServeHTTP(w, r)
}

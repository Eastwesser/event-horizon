package balancer

import (
	"log/slog"
	"math/rand/v2"
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
    // Mutex       sync.Mutex
    Healthy     atomic.Bool // set by the health-check goroutine in app
    Proxy       *httputil.ReverseProxy
}

type LeastConnBalancer struct {
	backends []*Backend
	mu       sync.RWMutex
	log      *slog.Logger
}

// func NewLeastConnBalancer(urls []string) *LeastConnBalancer {
//     backends := make([]*Backend, 0, len(urls))

//     for _, rawURL := range urls {
//         u, err := url.Parse(rawURL)
//         if err != nil {
//             log.Printf("Failed to parse URL %s: %v", rawURL, err)
//             continue
//         }

//         backend := &Backend{
//             URL:   u,
//             Proxy: httputil.NewSingleHostReverseProxy(u),
//         }

//         // custom proxy for logging
//         backend.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
//             log.Printf("Proxy error: %v", err)
//             http.Error(w, "Backend error", http.StatusBadGateway)
//         }

//         backends = append(backends, backend)
//     }

//     return &LeastConnBalancer{
//         backends: backends,
//     }
// }
func NewLeastConnBalancer(urls []string, log *slog.Logger) *LeastConnBalancer {
	if log == nil {
		log = slog.Default()
	}
	backends := make([]*Backend, 0, len(urls))

	for _, rawURL := range urls {
		u, err := url.Parse(rawURL)
		if err != nil {
			log.Error("failed to parse backend URL", "url", rawURL, "err", err)
			continue
		}

		backend := &Backend{
			URL:   u,
			Proxy: httputil.NewSingleHostReverseProxy(u),
		}
		backend.Healthy.Store(true) // optimistic: assume healthy until first check fails

		b := backend // capture for closure
		b.Proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Warn("proxy error", "backend", b.URL.Host, "method", r.Method, "path", r.URL.Path, "err", err)
			http.Error(w, "Backend error", http.StatusBadGateway)
		}

		backends = append(backends, backend)
	}

	return &LeastConnBalancer{
		backends: backends,
		log:      log,
	}
}

// func (lb *LeastConnBalancer) getLeastConnBackend() *Backend {
//     lb.mu.RLock()
//     defer lb.mu.RUnlock()

//     if len(lb.backends) == 0 {
//         return nil
//     }

//     var selected *Backend
//     var minConns int32 = 2147483647 // MaxInt32

//     for _, b := range lb.backends {
//         conns := atomic.LoadInt32(&b.ActiveConns)
//         if conns < minConns {
//             minConns = conns
//             selected = b
//         }
//     }

//     return selected
// }

// Backends exposes the slice for the health-check goroutine.
func (lb *LeastConnBalancer) Backends() []*Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.backends
}

// getLeastConnBackend picks the least-loaded healthy backend.
func (lb *LeastConnBalancer) getLeastConnBackend() *Backend {
	return lb.getLeastConnBackendExcluding(nil)
}

// getLeastConnBackendExcluding skips healthy=false and any backend already in `skip`.
// When several backends share the minimum ActiveConns, pick one at random (jitter)
// so sequential traffic is spread across the pool instead of always hitting the first.
func (lb *LeastConnBalancer) getLeastConnBackendExcluding(skip map[*Backend]struct{}) *Backend {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if len(lb.backends) == 0 {
		return nil
	}

	candidates := make([]*Backend, 0, len(lb.backends))
	var minConns int32 = 2147483647 // MaxInt32

	for _, b := range lb.backends {
		if !b.Healthy.Load() {
			continue
		}
		if skip != nil {
			if _, ok := skip[b]; ok {
				continue
			}
		}
		conns := atomic.LoadInt32(&b.ActiveConns)
		switch {
		case conns < minConns:
			minConns = conns
			candidates = candidates[:0]
			candidates = append(candidates, b)
		case conns == minConns:
			candidates = append(candidates, b)
		}
	}

	if len(candidates) == 0 {
		return nil
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	return candidates[rand.IntN(len(candidates))]
}

// func (lb *LeastConnBalancer) getTotalActiveConns() int32 {
//     lb.mu.RLock()
//     defer lb.mu.RUnlock()
    
//     var total int32
//     for _, b := range lb.backends {
//         total += atomic.LoadInt32(&b.ActiveConns)
//     }
//     return total
// }
func (lb *LeastConnBalancer) getTotalActiveConns() int32 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var total int32
	for _, b := range lb.backends {
		total += atomic.LoadInt32(&b.ActiveConns)
	}
	return total
}

// HealthyBackends returns how many backends are currently healthy (for /ready).
func (lb *LeastConnBalancer) HealthyBackends() int {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	var n int
	for _, b := range lb.backends {
		if b.Healthy.Load() {
			n++
		}
	}
	return n
}

func (lb *LeastConnBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics.RequestsTotal.Inc()

	backend := lb.getLeastConnBackend()

	if backend == nil {
		w.Header().Set("Retry-After", "5")
		http.Error(w, "No available backends", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt32(&backend.ActiveConns, 1)
	metrics.BackendActiveConnections.WithLabelValues(backend.URL.Host).Set(float64(atomic.LoadInt32(&backend.ActiveConns)))
	metrics.ActiveConnections.Set(float64(lb.getTotalActiveConns()))
	defer func() {
		n := atomic.AddInt32(&backend.ActiveConns, -1)
		metrics.BackendActiveConnections.WithLabelValues(backend.URL.Host).Set(float64(n))
		metrics.ActiveConnections.Set(float64(lb.getTotalActiveConns()))
	}()

	lb.log.Debug("proxying", "backend", backend.URL.Host, "method", r.Method, "path", r.URL.Path, "active", atomic.LoadInt32(&backend.ActiveConns))

	// Response header (debug): which backend served the request.
	w.Header().Set("X-Balancer-Backend", backend.URL.Host)

	backend.Proxy.ServeHTTP(w, r)
}

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    ActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "balancer_active_connections",
            Help: "Total active connections across all backends",
        },
    )

    RequestsTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "balancer_requests_total",
            Help: "Total number of requests processed by balancer",
        },
    )
)

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Week 8 business metrics (shop + fulfillment).
var (
	OrdersTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_total",
			Help: "Total successful shop purchases",
		},
	)
	OrdersRevenueTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_revenue_total",
			Help: "Sum of item prices for successful shop purchases",
		},
	)
	PurchasesFulfilledTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "purchases_fulfilled_total",
			Help: "purchase.fulfilled events published by fulfillment",
		},
	)
	PurchaseRefundFailures = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "shop_purchase_refund_failures_total",
			Help: "Shop could not refund tickets immediately after a failed purchase; outbox will retry",
		},
	)
	AssemblyDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "assembly_duration_seconds",
			Help:    "Time from purchase.paid to purchase.fulfilled",
			Buckets: []float64{0.5, 1, 2, 5, 10, 15, 30, 60, 120},
		},
	)
)

// RecordOrder increments shop business counters.
func RecordOrder(price float64) {
	OrdersTotal.Inc()
	if price > 0 {
		OrdersRevenueTotal.Add(price)
	}
}

// ObserveAssembly records fulfillment assembly duration.
func ObserveAssembly(seconds float64) {
	AssemblyDurationSeconds.Observe(seconds)
}

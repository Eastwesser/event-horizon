package config

import "os"

type Config struct {
    // HTTP
    Port        string
    MetricsPort string

    // Services
    AuthAddr        string
    GameAddr        string
    BillingAddr     string
    LeaderboardAddr string
    ProfileAddr     string
    ShopAddr        string
    InventoryAddr   string
    PaymentAddr     string
    AuthorsAddr     string
    HistoryAddr     string
    AnalyticsAddr   string

    // Infrastructure
    NATSUrl   string
    RedisAddr string

    // Payment webhook (Boosty / manual confirm)
    PaymentWebhookSecret string
}

func Load() *Config {
    return &Config{
        Port:            getEnv("PORT", "8080"),
        MetricsPort:     getEnv("METRICS_PORT", "9095"),
        AuthAddr:        getEnv("AUTH_ADDR", "localhost:50051"),
        GameAddr:        getEnv("GAME_ADDR", "localhost:50052"),
        BillingAddr:     getEnv("BILLING_ADDR", "localhost:50053"),
        LeaderboardAddr: getEnv("LEADERBOARD_ADDR", "localhost:50054"),
        ProfileAddr:     getEnv("PROFILE_ADDR", "profile:50060"),
        ShopAddr:        getEnv("SHOP_ADDR", "shop:50055"),
        InventoryAddr:   getEnv("INVENTORY_ADDR", "inventory:50059"),
        PaymentAddr:     getEnv("PAYMENT_ADDR", "payment:50058"),
        AuthorsAddr:     getEnv("AUTHORS_ADDR", "authors:50061"),
        HistoryAddr:     getEnv("HISTORY_ADDR", "history:50062"),
        AnalyticsAddr:   getEnv("ANALYTICS_ADDR", "analytics:50057"),
        NATSUrl:              getEnv("NATS_URL", "nats://localhost:4222"),
        RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
        PaymentWebhookSecret: getEnv("PAYMENT_WEBHOOK_SECRET", ""),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

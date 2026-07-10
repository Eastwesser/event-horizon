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

    // Infrastructure
    NATSUrl   string
    RedisAddr string
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
        NATSUrl:         getEnv("NATS_URL", "nats://localhost:4222"),
        RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

package config

import (
    "os"
    "strconv"
)

type Config struct {
    GRPCPort     string
    MetricsPort  string
    DBHost       string
    DBPort       string
    DBUser       string
    DBPassword   string
    DBName       string
    RedisAddr    string
    RedisDB      int
    NATSUrl      string
    BillingAddr  string
    PaymentAddr  string
}

func Load() *Config {
    return &Config{
        GRPCPort:    getEnv("GRPC_PORT", "50055"),
        MetricsPort: getEnv("METRICS_PORT", "9095"),
        DBHost:      getEnv("DB_HOST", "localhost"),
        DBPort:      getEnv("DB_PORT", "5465"),
        DBUser:      getEnv("DB_USER", "eventhorizon"),
        DBPassword:  getEnv("DB_PASSWORD", "eventhorizon"),
        DBName:      getEnv("DB_NAME", "eventhorizon_shop"),
        RedisAddr:   getEnv("REDIS_ADDR", "localhost:6383"),
        RedisDB:     getEnvAsInt("REDIS_DB", 0),
        NATSUrl:     getEnv("NATS_URL", "nats://localhost:4222"),
        BillingAddr: getEnv("BILLING_ADDR", "billing:50053"),
        PaymentAddr: getEnv("PAYMENT_ADDR", "payment:50058"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}
package config

import (
    "os"
    "strconv"
)

type Config struct {
    GRPCPort    string
    RedisAddr   string
    RedisDB     int
    NATSUrl     string
}

func Load() *Config {
    return &Config{
        GRPCPort:  getEnv("GRPC_PORT", "50054"),
        RedisAddr: getEnv("REDIS_ADDR", "127.0.0.1:6382"),
        RedisDB:   getEnvAsInt("REDIS_DB", 0),
        NATSUrl:   getEnv("NATS_URL", "nats://localhost:4222"),
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

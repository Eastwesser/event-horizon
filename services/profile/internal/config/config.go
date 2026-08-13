package config

import (
    "os"
    "strconv"
)

type Config struct {
    GRPCPort           string
    DBHost             string
    DBPort             string
    DBUser             string
    DBPassword         string
    DBName             string
    NATSUrl            string
    MetricsPort        string
    RedisAddr          string
    ProfileCacheTTLMin int
}

func Load() *Config {
    return &Config{
        GRPCPort:           getEnv("GRPC_PORT", "50060"),
        DBHost:             getEnv("DB_HOST", "localhost"),
        DBPort:             getEnv("DB_PORT", "5464"),
        DBUser:             getEnv("DB_USER", "eventhorizon"),
        DBPassword:         getEnv("DB_PASSWORD", "eventhorizon"),
        DBName:             getEnv("DB_NAME", "eventhorizon_profile"),
        NATSUrl:            getEnv("NATS_URL", "nats://localhost:4222"),
        MetricsPort:        getEnv("METRICS_PORT", "9099"),
        RedisAddr:          getEnv("REDIS_ADDR", "localhost:6385"),
        ProfileCacheTTLMin: getEnvAsInt("PROFILE_CACHE_TTL_MINUTES", 5),
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

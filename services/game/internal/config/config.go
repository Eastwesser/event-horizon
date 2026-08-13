package config

import (
    "os"
    "strconv"
)

type Config struct {
    GRPCPort    string
    MetricsPort string
    DBHost      string
    DBPort      string
    DBUser      string
    DBPassword  string
    DBName      string
    NATSUrl     string
}

func Load() *Config {
    return &Config{
        GRPCPort:    getEnv("GRPC_PORT", "50052"),
        MetricsPort: getEnv("METRICS_PORT", "9092"),
        DBHost:      getEnv("DB_HOST", "localhost"),
        DBPort:      getEnv("DB_PORT", "5461"),
        DBUser:      getEnv("DB_USER", "eventhorizon"),
        DBPassword:  getEnv("DB_PASSWORD", "eventhorizon"),
        DBName:      getEnv("DB_NAME", "eventhorizon_game"),
        NATSUrl:     getEnv("NATS_URL", "nats://localhost:4222"),
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

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
    DBHost       string
    DBPort       string
    DBUser       string
    DBPassword   string
    DBName       string
    GameDBHost     string
    GameDBPort     string
    GameDBUser     string
    GameDBPassword string
    GameDBName     string
}

func Load() *Config {
    return &Config{
        GRPCPort:     getEnv("GRPC_PORT", "50054"),
        RedisAddr:    getEnv("REDIS_ADDR", "127.0.0.1:6382"),
        RedisDB:      getEnvAsInt("REDIS_DB", 0),
        NATSUrl:      getEnv("NATS_URL", "nats://localhost:4222"),
        DBHost:       getEnv("DB_HOST", "localhost"),
        DBPort:       getEnv("DB_PORT", "5463"),
        DBUser:       getEnv("DB_USER", "eventhorizon"),
        DBPassword:   getEnv("DB_PASSWORD", "eventhorizon"),
        DBName:       getEnv("DB_NAME", "eventhorizon_game"),
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

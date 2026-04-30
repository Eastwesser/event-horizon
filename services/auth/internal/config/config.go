package config

import (
    "os"
    "strconv"
)

type Config struct {
    GRPCPort    string
    DBHost      string
    DBPort      string
    DBUser      string
    DBPassword  string
    DBName      string
    JWTSecret   string
    JWTExpHours int
}

func Load() *Config {
    return &Config{
        GRPCPort:    getEnv("GRPC_PORT", "50051"),
        DBHost:      getEnv("DB_HOST", "localhost"),
        DBPort:      getEnv("DB_PORT", "5460"),
        DBUser:      getEnv("DB_USER", "eventhorizon"),
        DBPassword:  getEnv("DB_PASSWORD", "eventhorizon"),
        DBName:      getEnv("DB_NAME", "eventhorizon"),
        JWTSecret:   getEnv("JWT_SECRET", "your-secret-key-change-me"),
        JWTExpHours: getEnvAsInt("JWT_EXP_HOURS", 24),
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

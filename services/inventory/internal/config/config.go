package config

import (
    "fmt"
    "os"
    "strconv"
)

type Config struct {
    GRPCPort    int
    MetricsPort int

    // PostgreSQL
    PGHost     string
    PGPort     int
    PGUser     string
    PGPassword string
    PGDBName   string

    // MongoDB
    MongoURI    string
    MongoDBName string

    // Redis
    RedisAddr string

    // Драйвер: "postgres" или "mongo"
    Driver string

    // NATS
    NATSURL string
}

func Load() (*Config, error) {
    cfg := &Config{
        GRPCPort:    getEnvAsInt("INVENTORY_GRPC_PORT", 50059),
        MetricsPort: getEnvAsInt("INVENTORY_METRICS_PORT", 9096),

        PGHost:     getEnv("INVENTORY_PG_HOST", "localhost"),
        PGPort:     getEnvAsInt("INVENTORY_PG_PORT", 5466),
        PGUser:     getEnv("INVENTORY_PG_USER", "postgres"),
        PGPassword: getEnv("INVENTORY_PG_PASSWORD", "postgres"),
        PGDBName:   getEnv("INVENTORY_PG_DB", "inventory"),

        MongoURI:    getEnv("INVENTORY_MONGO_URI", "mongodb://localhost:27017"),
        MongoDBName: getEnv("INVENTORY_MONGO_DB", "inventory"),

        RedisAddr: getEnv("INVENTORY_REDIS_ADDR", "localhost:6383"),

        Driver:  getEnv("INVENTORY_DRIVER", "postgres"),
        NATSURL: getEnv("NATS_URL", "nats://localhost:4222"),
    }

    return cfg, nil
}

func (c *Config) PGDSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        c.PGHost, c.PGPort, c.PGUser, c.PGPassword, c.PGDBName,
    )
}

func getEnv(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists && value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value, exists := os.LookupEnv(key); exists && value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}
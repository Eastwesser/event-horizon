package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GRPCPort    string
	MetricsPort string

	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	RedisAddr string

	NATSURL string

	// BoostyCheckoutURL is the public Boosty page (or stub) users are redirected to.
	BoostyCheckoutURL string
	// WebhookSecret must match ConfirmPayment.webhook_secret when set (non-empty).
	WebhookSecret string
	// SubscriptionDays is how long an activated plan lasts.
	SubscriptionDays int

	LogLevel  string
	LogFormat string
}

func Load() *Config {
	days, _ := strconv.Atoi(getenv("SUBSCRIPTION_DAYS", "30"))
	if days <= 0 {
		days = 30
	}
	return &Config{
		GRPCPort:          getenv("GRPC_PORT", "50058"),
		MetricsPort:       getenv("METRICS_PORT", "9103"),
		DBHost:            getenv("DB_HOST", "localhost"),
		DBPort:            getenv("DB_PORT", "5467"),
		DBUser:            getenv("DB_USER", "eventhorizon"),
		DBPass:            getenv("DB_PASSWORD", "eventhorizon"),
		DBName:            getenv("DB_NAME", "eventhorizon_payment"),
		RedisAddr:         getenv("REDIS_ADDR", "localhost:6386"),
		NATSURL:           getenv("NATS_URL", "nats://localhost:4222"),
		BoostyCheckoutURL: getenv("BOOSTY_CHECKOUT_URL", "https://boosty.to/eastwesser"),
		WebhookSecret:     getenv("PAYMENT_WEBHOOK_SECRET", ""),
		SubscriptionDays:  days,
		LogLevel:          getenv("LOG_LEVEL", "info"),
		LogFormat:         getenv("LOG_FORMAT", "text"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName,
	)
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

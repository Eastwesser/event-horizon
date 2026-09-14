package config

import (
	"github.com/wb-go/wbf/config"
)

type Config struct {
	Port          string
	StoragePath   string
	KafkaBrokers  []string
	KafkaTopic    string
	KafkaGroupID  string
}

func Load() *Config {
	cfg := config.New()
	cfg.SetDefault("port", "8080")
	cfg.SetDefault("storage_path", "./storage")
	cfg.SetDefault("kafka_brokers", []string{"localhost:9092"})
	cfg.SetDefault("kafka_topic", "image-processing")
	cfg.SetDefault("kafka_group_id", "image-processor-group")

	cfg.EnableEnv("")
	cfg.LoadEnvFiles(".env")

	kafkaBrokers := cfg.GetStringSlice("kafka_brokers")
	if len(kafkaBrokers) == 0 {
		kafkaBrokers = []string{"localhost:9092"}
	}

	port := cfg.GetString("port")
	if port == "" {
		port = "8080"
	}

	storagePath := cfg.GetString("storage_path")
	if storagePath == "" {
		storagePath = "./storage"
	}

	kafkaTopic := cfg.GetString("kafka_topic")
	if kafkaTopic == "" {
		kafkaTopic = "image-processing"
	}

	kafkaGroupID := cfg.GetString("kafka_group_id")
	if kafkaGroupID == "" {
		kafkaGroupID = "image-processor-group"
	}

	return &Config{
		Port:         port,
		StoragePath:  storagePath,
		KafkaBrokers: kafkaBrokers,
		KafkaTopic:   kafkaTopic,
		KafkaGroupID: kafkaGroupID,
	}
}


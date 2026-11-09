package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server        ServerConfig
	Auth          AuthConfig
	Elasticsearch ESConfig
	Kafka         KafkaConfig
}

type ServerConfig struct {
	Port         string
	MaxBodyBytes int64
}

type AuthConfig struct {
	// Disabled turns off authentication entirely. Only for local development.
	Disabled  bool
	JWTSecret string
}

type ESConfig struct {
	URL         string
	Username    string
	Password    string
	IndexPrefix string
}

type KafkaConfig struct {
	Brokers       []string
	Topic         string
	DLQTopic      string
	ResultsTopic  string
	ConsumerGroup string
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvBool(key string, def bool) bool {
	v, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return def
	}
	return v
}

func getEnvInt64(key string, def int64) int64 {
	v, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil {
		return def
	}
	return v
}

// splitList parses a comma-separated list, e.g. "kafka1:9092, kafka2:9092".
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}

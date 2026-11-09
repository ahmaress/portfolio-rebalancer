package config

import (
	"errors"
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

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			MaxBodyBytes: getEnvInt64("MAX_BODY_BYTES", 1<<20),
		},
		Auth: AuthConfig{
			Disabled:  getEnvBool("AUTH_DISABLED", false),
			JWTSecret: os.Getenv("JWT_SECRET"),
		},
		Elasticsearch: ESConfig{
			URL:         getEnv("ELASTICSEARCH_URL", "http://localhost:9200"),
			Username:    os.Getenv("ELASTICSEARCH_USERNAME"),
			Password:    os.Getenv("ELASTICSEARCH_PASSWORD"),
			IndexPrefix: os.Getenv("ELASTICSEARCH_INDEX_PREFIX"),
		},
		Kafka: KafkaConfig{
			Brokers:       splitList(getEnv("KAFKA_BROKERS", "localhost:9092")),
			Topic:         getEnv("KAFKA_TOPIC", "rebalance-events"),
			DLQTopic:      getEnv("KAFKA_DLQ_TOPIC", "rebalance-events-dlq"),
			ResultsTopic:  getEnv("KAFKA_RESULTS_TOPIC", "rebalance-results"),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "portfolio-rebalancer"),
		},
	}
	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	var errs []error
	if !c.Auth.Disabled && len(c.Auth.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 characters (or set AUTH_DISABLED=true for local development)"))
	}
	if len(c.Kafka.Brokers) == 0 {
		errs = append(errs, errors.New("KAFKA_BROKERS must list at least one broker"))
	}
	if c.Server.MaxBodyBytes <= 0 {
		errs = append(errs, errors.New("MAX_BODY_BYTES must be positive"))
	}
	return errors.Join(errs...)
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

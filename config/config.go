package config

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

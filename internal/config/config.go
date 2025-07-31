package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Blockchain BlockchainConfig
	CORS       CORSConfig
	RateLimit  RateLimitConfig
	Logging    LoggingConfig
}

type ServerConfig struct {
	Port       string
	Env        string
	APIVersion string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type BlockchainConfig struct {
	LiskRPCURL           string
	LiskChainID          int64
	LiskTestnetRPCURL    string
	LiskTestnetChainID   int64
	SurveyContractAddr   string
	RewardContractAddr   string
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
}

type LoggingConfig struct {
	Level  string
	Format string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:       getEnv("PORT", "8080"),
			Env:        getEnv("ENV", "development"),
			APIVersion: getEnv("API_VERSION", "v1"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "survey2earn"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "survey2earn_db"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "change-this-secret-key"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
		Blockchain: BlockchainConfig{
			LiskRPCURL:           getEnv("LISK_RPC_URL", "https://rpc.api.lisk.com"),
			LiskChainID:          getEnvAsInt64("LISK_CHAIN_ID", 1135),
			LiskTestnetRPCURL:    getEnv("LISK_TESTNET_RPC_URL", "https://rpc.sepolia-api.lisk.com"),
			LiskTestnetChainID:   getEnvAsInt64("LISK_TESTNET_CHAIN_ID", 4202),
			SurveyContractAddr:   getEnv("SURVEY_CONTRACT_ADDRESS", ""),
			RewardContractAddr:   getEnv("REWARD_CONTRACT_ADDRESS", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins: strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
			AllowedMethods: strings.Split(getEnv("ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS"), ","),
			AllowedHeaders: strings.Split(getEnv("ALLOWED_HEADERS", "Content-Type,Authorization"), ","),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			Burst:             getEnvAsInt("RATE_LIMIT_BURST", 10),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return config, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetDatabaseDSN returns the PostgreSQL connection string
func (c *Config) GetDatabaseDSN() string {
	return "host=" + c.Database.Host +
		" port=" + c.Database.Port +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.Name +
		" sslmode=" + c.Database.SSLMode
}

// IsProduction checks if the environment is production
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// IsDevelopment checks if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}
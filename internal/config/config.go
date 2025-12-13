package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	OpenAI      OpenAIConfig
	Google      GoogleOAuthConfig
	VoiceVox    VoiceVoxConfig
	NCP_Storage NCloudStorageConfig
}

type NCloudStorageConfig struct {
	AccessKey  string
	SecretKey  string
	BucketName string
	Endpoint   string
}

type VoiceVoxConfig struct {
	VoiceVoxURL string
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type OpenAIConfig struct {
	APIKey string
	Model  string
}

type ServerConfig struct {
	Port string
	Mode string // debug, release, test
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "jptaku"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPEN_AI_API_KEY", ""),
			Model:  getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		},
		Google: GoogleOAuthConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:30001/api/auth/google/callback"),
		},
		VoiceVox: VoiceVoxConfig{
			VoiceVoxURL: getEnv("VOICEVOX_URL", "http://localhost:50021"),
		},
		NCP_Storage: NCloudStorageConfig{
			AccessKey:  getEnv("NCP_ACCESS_KEY", ""),
			SecretKey:  getEnv("NCP_SECRET_KEY", ""),
			BucketName: getEnv("NCP_BUCKET_NAME", ""),
			Endpoint:   getEnv("NCP_ENDPOINT", ""),
		},
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
		intValue, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: Invalid integer for %s, using default: %d", key, defaultValue)
			return defaultValue
		}
		return intValue
	}
	return defaultValue
}

package config

import "os"

type Config struct {
	Port        string
	MongoURI    string
	MongoDBName string

	VisionLLMBaseURL string
	VisionLLMAPIKey  string
	VisionLLMModel   string

	CoderLLMBaseURL string
	CoderLLMAPIKey  string
	CoderLLMModel   string
}

const defaultOpenRouterBaseURL = "https://openrouter.ai/api/v1"

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		MongoURI:    getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName: getEnv("MONGO_DB_NAME", "mongo_hack"),

		VisionLLMBaseURL: getEnv("VISION_LLM_BASE_URL", defaultOpenRouterBaseURL),
		VisionLLMAPIKey:  getEnv("VISION_LLM_API_KEY", ""),
		VisionLLMModel:   getEnv("VISION_LLM_MODEL", ""),

		CoderLLMBaseURL: getEnv("CODER_LLM_BASE_URL", defaultOpenRouterBaseURL),
		CoderLLMAPIKey:  getEnv("CODER_LLM_API_KEY", ""),
		CoderLLMModel:   getEnv("CODER_LLM_MODEL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

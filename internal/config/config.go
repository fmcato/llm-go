package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the configuration for the LLM client
type Config struct {
	APIKey       string
	BaseURL      string
	Model        string
	Temperature  float64
	SystemPrompt string
}

// LoadConfig loads configuration from environment variables
func LoadConfig(systemPrompt string) Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Warning: OPENAI_API_KEY environment variable is not set")
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	// Use the provided system prompt instead of environment variable
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}

	// Parse temperature with default value of 0.7
	temperatureStr := os.Getenv("OPENAI_TEMPERATURE")
	temperature := 0.7 // default temperature
	if temperatureStr != "" {
		if parsedTemp, err := strconv.ParseFloat(temperatureStr, 64); err == nil {
			// Validate temperature range (0.0 to 2.0)
			if parsedTemp >= 0.0 && parsedTemp <= 2.0 {
				temperature = parsedTemp
			} else {
				fmt.Printf("Warning: Temperature value %f is outside valid range (0.0-2.0), using default 0.7\n", parsedTemp)
			}
		} else {
			fmt.Printf("Warning: Invalid temperature value '%s', using default 0.7\n", temperatureStr)
		}
	}

	return Config{
		APIKey:       apiKey,
		BaseURL:      baseURL,
		Model:        model,
		Temperature:  temperature,
		SystemPrompt: systemPrompt,
	}
}

// ReadSystemPrompt reads the system prompt from a file
func ReadSystemPrompt(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read system prompt file: %w", err)
	}
	return strings.TrimSpace(string(content)), nil
}

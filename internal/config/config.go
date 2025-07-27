package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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

// LoadConfig loads configuration with CLI arguments taking precedence over environment variables
func LoadConfig(systemPrompt, cliModel string, cliTemperature float64) Config {
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

	// Prioritize CLI model over environment variable
	model := cliModel
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
		if model == "" {
			model = "gpt-4o"
		}
	}

	// Use the provided system prompt instead of environment variable
	if systemPrompt == "" {
		systemPrompt = "You are a helpful assistant."
	}

	// Prioritize CLI temperature over environment variable
	temperature := 0.7 // default temperature
	if cliTemperature != 0.0 {
		// Validate temperature range (0.0 to 2.0)
		if cliTemperature >= 0.0 && cliTemperature <= 2.0 {
			temperature = cliTemperature
		} else {
			fmt.Printf("Warning: Temperature value %f is outside valid range (0.0-2.0), using default 0.7\n", cliTemperature)
		}
	} else {
		// Fall back to environment variable
		temperatureStr := os.Getenv("OPENAI_TEMPERATURE")
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
	}

	return Config{
		APIKey:       apiKey,
		BaseURL:      baseURL,
		Model:        model,
		Temperature:  temperature,
		SystemPrompt: systemPrompt,
	}
}

// formatCurrentDateTime returns current datetime in "Tuesday 1 September 2025, 10:17 AM" format
func formatCurrentDateTime() string {
	now := time.Now()
	// Format day with ordinal suffix
	day := now.Day()

	// Format the complete string with weekday
	return fmt.Sprintf("%s %d %s %d, %s",
		now.Weekday().String(),
		day,
		now.Month().String(),
		now.Year(),
		now.Format("3:04 PM"))
}

// ReadSystemPrompt reads the system prompt from a file
func ReadSystemPrompt(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read system prompt file: %w", err)
	}
	prompt := strings.TrimSpace(string(content))
	return strings.ReplaceAll(prompt, "{{currentDateTime}}", formatCurrentDateTime()), nil
}

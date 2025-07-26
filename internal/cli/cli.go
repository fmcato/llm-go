package cli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// CLI handles command-line interface operations
type CLI struct {
	hideThinking  bool
	model         string
	temperature   float64
	outputJson    bool
	showModelInfo bool
	reader        *bufio.Reader
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		reader: bufio.NewReader(os.Stdin),
	}
}

// ParseFlags parses command-line flags
func (c *CLI) ParseFlags() {
	flag.BoolVar(&c.hideThinking, "hide-thinking", false, "Hide thinking/reasoning parts of the response")
	flag.StringVar(&c.model, "model", "", "Model to use for completions")
	flag.Float64Var(&c.temperature, "temperature", 0.0, "Temperature for completions (0.0-2.0)")
	flag.BoolVar(&c.outputJson, "json", false, "Output response as JSON")
	flag.BoolVar(&c.showModelInfo, "model-info", false, "Display detailed model information")
	flag.Parse()
}

// GetHideThinking returns the hide-thinking flag value
func (c *CLI) GetHideThinking() bool {
	return c.hideThinking
}

// GetModel returns the model flag value
func (c *CLI) GetModel() string {
	return c.model
}

// GetTemperature returns the temperature flag value
func (c *CLI) GetTemperature() float64 {
	return c.temperature
}

// GetJSON returns the json flag value
func (c *CLI) GetJSON() bool {
	return c.outputJson
}

// GetSystemPromptFile returns the system prompt file path from command-line arguments
func (c *CLI) GetSystemPromptFile() string {
	if flag.NArg() < 1 {
		return ""
	}
	return flag.Arg(0)
}

// ShowUsage displays usage information
func (c *CLI) ShowUsage() {
	fmt.Println("Usage: llm-go [options] <system-prompt-file>")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  OPENAI_API_KEY      API key for OpenAI-compatible API")
	fmt.Println("  OPENAI_BASE_URL     Base URL for OpenAI-compatible API (default: https://api.openai.com/v1)")
	fmt.Println("  OPENAI_MODEL        Model to use for completions (default: gpt-4o)")
	fmt.Println("  OPENAI_TEMPERATURE  Temperature for completions (0.0-2.0, default: 0.7)")
}

// GetUserInput gets input from the user
func (c *CLI) GetUserInput() (string, error) {
	message, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return strings.TrimSpace(message), nil
}

// ShowError displays an error message
func (c *CLI) ShowError(err error) {
	fmt.Printf("Error: %v\n", err)
}

// ShouldQuit checks if the user wants to quit
func (c *CLI) ShouldQuit(message string) bool {
	return message == "/quit"
}

// IsValidMessage checks if the message is valid (not empty)
func (c *CLI) IsValidMessage(message string) bool {
	return message != ""
}

// GetShowModelInfo returns the model-info flag value
func (c *CLI) GetShowModelInfo() bool {
	return c.showModelInfo
}

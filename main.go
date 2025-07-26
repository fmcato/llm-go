package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"llm-go/internal/cli"
	"llm-go/internal/config"
	"llm-go/internal/llm"
	"llm-go/internal/memory"
)

const (
	// Do not change
	startThinkTag = "<think>"
	endThinkTag   = "</think>"
)

// removeThinkingBlocks removes thinking blocks (including tags and content) from responses
// and returns only the actual response content after the thinking block
func removeThinkingBlocks(s string) string {

	startIdx := strings.Index(s, startThinkTag)
	if startIdx == -1 {
		return s // No thinking block found, return original
	}

	// Find the end of the thinking block
	afterStart := s[startIdx+len(startThinkTag):]
	endIdx := strings.Index(afterStart, endThinkTag)
	if endIdx == -1 {
		return s // No end tag found, return original
	}

	// Calculate position after the thinking block
	afterEnd := startIdx + len(startThinkTag) + endIdx + len(endThinkTag)

	// Return only content after the thinking block, trimmed
	return strings.TrimSpace(s[afterEnd:])
}

// extractThinkingBlocks extracts thinking blocks (including tags and content) from responses
func extractThinkingBlocks(s string) string {
	startIdx := strings.Index(s, startThinkTag)
	if startIdx == -1 {
		return "" // No thinking block found
	}

	// Find the end of the thinking block
	afterStart := s[startIdx+len(startThinkTag):]
	endIdx := strings.Index(afterStart, endThinkTag)
	if endIdx == -1 {
		return "" // No end tag found
	}

	// Calculate position after the thinking block
	afterEnd := startIdx + len(startThinkTag) + endIdx + len(endThinkTag)

	// Return only the thinking block content (including tags)
	return strings.TrimSpace(s[startIdx:afterEnd])
}

func main() {
	cliHandler := initCLI()

	// Handle model info display
	if cliHandler.GetShowModelInfo() {
		cfg := loadConfig(cliHandler)
		client := initLLMClient(cfg)
		if err := client.DisplayModelInfo(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cfg := loadConfig(cliHandler)
	client := initLLMClient(cfg)
	mem := initMemory(cfg)
	runConversationLoop(cliHandler, client, mem)
}

// initCLI initializes and parses command line flags
func initCLI() *cli.CLI {
	cliHandler := cli.NewCLI()
	cliHandler.ParseFlags()
	return cliHandler
}

// loadConfig validates inputs and loads configuration
func loadConfig(cliHandler *cli.CLI) *config.Config {
	var systemPrompt string

	// Check if system prompt file path is provided as argument
	systemPromptFile := cliHandler.GetSystemPromptFile()
	if systemPromptFile != "" {
		// Read system prompt from file
		var err error
		systemPrompt, err = config.ReadSystemPrompt(systemPromptFile)
		if err != nil {
			cliHandler.ShowError(err)
			os.Exit(1)
		}
	}
	// If no system prompt file is provided, systemPrompt remains empty

	// Load configuration with system prompt, model, and temperature
	cfg := config.LoadConfig(systemPrompt, cliHandler.GetModel(), cliHandler.GetTemperature())

	// Validate API key
	if cfg.APIKey == "" {
		cliHandler.ShowError(nil)
		os.Exit(1)
	}

	return &cfg
}

// initLLMClient creates and configures the LLM client
func initLLMClient(cfg *config.Config) *llm.Client {
	// Create LLM client
	llmConfig := llm.Config{
		APIKey:       cfg.APIKey,
		BaseURL:      cfg.BaseURL,
		Model:        cfg.Model,
		Temperature:  cfg.Temperature,
		SystemPrompt: cfg.SystemPrompt,
	}
	return llm.NewClient(llmConfig)
}

// initMemory initializes conversation history with system message
func initMemory(cfg *config.Config) *memory.Memory {
	// Create memory for conversation history
	mem := memory.NewMemory()
	// Initialize conversation history with system message if provided
	if cfg.SystemPrompt != "" {
		mem.AddSystemMessage(cfg.SystemPrompt)
	}
	return mem
}

// runConversationLoop handles the main conversation interaction
func runConversationLoop(cliHandler *cli.CLI, client *llm.Client, mem *memory.Memory) {
	for {
		message, shouldExit := handleUserInput(cliHandler)
		if shouldExit {
			if !cliHandler.GetJSON() {
				client.DisplayTotalUsage()
			}
			return
		}

		// Skip empty messages
		if !cliHandler.IsValidMessage(message) {
			continue
		}

		// Add user message to history
		mem.AddUserMessage(message)

		response, err := processResponse(cliHandler, client, mem)
		if err != nil {
			cliHandler.ShowError(err)
			continue
		}

		displayResults(cliHandler, client, response)

		// Add assistant response to history (without thinking blocks)
		mem.AddAssistantMessage(removeThinkingBlocks(response))
	}
}

// handleUserInput gets and validates user input
func handleUserInput(cliHandler *cli.CLI) (string, bool) {
	// Get user input
	if !cliHandler.GetJSON() {
		fmt.Print("\nEnter your message (or '/quit' to exit): ")
	}
	message, err := cliHandler.GetUserInput()
	if err != nil {
		// Check if it's EOF (end of input) - treat as quit signal
		if errors.Is(err, io.EOF) {
			return "", true
		}
		cliHandler.ShowError(err)
		return "", false
	}

	// Check for quit command
	if cliHandler.ShouldQuit(message) {
		return message, true
	}

	return message, false
}

// processResponse handles streaming and processing of LLM responses
func processResponse(cliHandler *cli.CLI, client *llm.Client, mem *memory.Memory) (string, error) {
	// Send message and stream response
	chunkChan := make(chan string)
	resultChan := make(chan struct {
		response string
		err      error
	}, 1)

	// Only show "Response:" header in non-JSON mode
	if !cliHandler.GetJSON() {
		fmt.Println("\nResponse:")
	}

	// Start streaming in a goroutine
	go func() {
		response, err := client.StreamResponse(mem.GetMessages(), cliHandler.GetHideThinking(), chunkChan)
		resultChan <- struct {
			response string
			err      error
		}{response: response, err: err}
	}()

	// Print chunks as they arrive (only in non-JSON mode)
	for chunk := range chunkChan {
		if !cliHandler.GetJSON() {
			fmt.Print(chunk)
		}
	}

	// Wait for streaming to complete and get result
	result := <-resultChan
	return result.response, result.err
}

// displayResults formats and displays the response based on output mode
func displayResults(cliHandler *cli.CLI, client *llm.Client, response string) {
	if !cliHandler.GetJSON() {
		client.DisplayTokenUsage()
		return
	}
	// Handle JSON output if requested
	stats := client.GetStats()
	jsonResponse := map[string]interface{}{
		"response": removeThinkingBlocks(response),
		"thinking": extractThinkingBlocks(response),
		"stats": map[string]interface{}{
			"tokens": map[string]int{
				"input":  stats.InputTokens,
				"output": stats.OutputTokens,
				"total":  stats.InputTokens + stats.OutputTokens,
			},
			"time": map[string]int64{
				"thinking_ms": stats.ThinkingTime.Milliseconds(),
				"response_ms": stats.ResponseTime.Milliseconds(),
				"total_ms":    (stats.ThinkingTime + stats.ResponseTime).Milliseconds(),
			},
		},
	}

	jsonData, err := json.Marshal(jsonResponse)
	if err != nil {
		cliHandler.ShowError(fmt.Errorf("error marshaling JSON: %w", err))
		return
	}
	fmt.Println(string(jsonData))
}

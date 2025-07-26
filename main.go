package main

import (
	"os"

	"llm-go/internal/cli"
	"llm-go/internal/config"
	"llm-go/internal/llm"
	"llm-go/internal/memory"
)

func main() {
	// Create CLI handler
	cliHandler := cli.NewCLI()
	cliHandler.ParseFlags()

	// Check if system prompt file path is provided as argument
	systemPromptFile := cliHandler.GetSystemPromptFile()
	if systemPromptFile == "" {
		cliHandler.ShowUsage()
		os.Exit(1)
	}

	// Read system prompt from file
	systemPrompt, err := config.ReadSystemPrompt(systemPromptFile)
	if err != nil {
		cliHandler.ShowError(err)
		os.Exit(1)
	}

	// Load configuration with system prompt
	cfg := config.LoadConfig(systemPrompt)

	// Validate API key
	if cfg.APIKey == "" {
		cliHandler.ShowError(nil)
		os.Exit(1)
	}

	// Create LLM client
	llmConfig := llm.Config{
		APIKey:       cfg.APIKey,
		BaseURL:      cfg.BaseURL,
		Model:        cfg.Model,
		Temperature:  cfg.Temperature,
		SystemPrompt: cfg.SystemPrompt,
	}
	client := llm.NewClient(llmConfig)

	// Create memory for conversation history
	mem := memory.NewMemory()

	// Initialize conversation history with system message
	mem.AddSystemMessage(cfg.SystemPrompt)

	// Continuous conversation loop
	for {
		// Get user input
		message, err := cliHandler.GetUserInput()
		if err != nil {
			cliHandler.ShowError(err)
			continue
		}

		// Check for quit command
		if cliHandler.ShouldQuit(message) {
			cliHandler.ShowGoodbye()
			client.DisplayTotalUsage()
			break
		}

		// Skip empty messages
		if !cliHandler.IsValidMessage(message) {
			continue
		}

		// Add user message to history
		mem.AddUserMessage(message)

		// Send message and stream response
		response, err := client.StreamResponse(mem.GetMessages(), cliHandler.GetHideThinking())
		if err != nil {
			cliHandler.ShowError(err)
			continue
		}

		// Display token usage for this interaction
		client.DisplayTokenUsage()

		// Add assistant response to history
		mem.AddAssistantMessage(response)
	}
}

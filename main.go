package main

import (
	"fmt"
	"os"
	"strings"

	"llm-go/internal/cli"
	"llm-go/internal/config"
	"llm-go/internal/llm"
	"llm-go/internal/memory"
)

// removeThinkingBlocks removes thinking blocks (including tags and content) from responses
// and returns only the actual response content after the thinking block
func removeThinkingBlocks(s string) string {
	startThinkTag := "<tool_call>"
	endThinkTag := "</tool_call>"

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
		chunkChan := make(chan string)
		resultChan := make(chan struct {
			response string
			err      error
		}, 1)

		fmt.Println("\nResponse:")

		// Start streaming in a goroutine
		go func() {
			response, err := client.StreamResponse(mem.GetMessages(), cliHandler.GetHideThinking(), chunkChan)
			resultChan <- struct {
				response string
				err      error
			}{response: response, err: err}
		}()

		// Print chunks as they arrive
		for chunk := range chunkChan {
			fmt.Print(chunk)
		}

		// Wait for streaming to complete and get result
		result := <-resultChan
		if result.err != nil {
			cliHandler.ShowError(result.err)
			continue
		}
		response := result.response

		// Display token usage for this interaction
		client.DisplayTokenUsage()

		// Add assistant response to history (without thinking blocks)
		mem.AddAssistantMessage(removeThinkingBlocks(response))
	}
}

package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

const (
	startThinkTag = "<think>"
	endThinkTag   = "</think>"
)

// Client wraps the OpenAI client with additional functionality
type Client struct {
	client *openai.Client
	config Config

	// Token tracking
	totalInputTokens    int
	totalOutputTokens   int
	currentInputTokens  int
	currentOutputTokens int
	mutex               sync.Mutex
}

// Config holds the configuration for the LLM client
type Config struct {
	APIKey       string
	BaseURL      string
	Model        string
	Temperature  float64
	SystemPrompt string
}

// NewClient creates a new LLM client with the given configuration
func NewClient(config Config) *Client {
	client := openai.NewClient(
		option.WithAPIKey(config.APIKey),
		option.WithBaseURL(config.BaseURL),
	)

	return &Client{
		client: &client,
		config: config,
	}
}

// DisplayTokenUsage shows the token usage for the current interaction
func (c *Client) DisplayTokenUsage() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	fmt.Printf("\nTokens: Input %d | Output %d | Total %d\n",
		c.currentInputTokens, c.currentOutputTokens,
		c.currentInputTokens+c.currentOutputTokens)
}

// DisplayTotalUsage shows the total token usage across all interactions
func (c *Client) DisplayTotalUsage() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	fmt.Printf("\nTotal tokens used: Input %d | Output %d | Combined %d\n",
		c.totalInputTokens, c.totalOutputTokens,
		c.totalInputTokens+c.totalOutputTokens)
}

// StreamResponse sends a message with conversation history and streams the response
// while concurrently sending chunks to the provided channel
func (c *Client) StreamResponse(messages []openai.ChatCompletionMessageParamUnion, hideThinking bool, chunkChan chan<- string) (string, error) {
	// Reset current interaction token counts
	c.mutex.Lock()
	c.currentInputTokens = 0
	c.currentOutputTokens = 0
	c.mutex.Unlock()

	// Create streaming chat completion with usage tracking
	stream := c.client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
		Model:       c.config.Model,
		Messages:    messages,
		Temperature: param.NewOpt(c.config.Temperature),
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: param.NewOpt(true),
		},
	})

	var fullResponse strings.Builder
	var inThinkingBlock bool

	for stream.Next() {
		chunk := stream.Current()

		// Check for usage data in the chunk
		if chunk.Usage.PromptTokens > 0 {
			c.mutex.Lock()
			c.currentInputTokens = int(chunk.Usage.PromptTokens)
			c.currentOutputTokens = int(chunk.Usage.CompletionTokens)
			c.totalInputTokens += c.currentInputTokens
			c.totalOutputTokens += c.currentOutputTokens
			c.mutex.Unlock()
		}

		if len(chunk.Choices) == 0 {
			continue
		}
		delta := chunk.Choices[0].Delta
		if delta.Content == "" {
			continue
		}
		text := delta.Content

		if !inThinkingBlock && text == startThinkTag {
			inThinkingBlock = true
		}
		if inThinkingBlock && text == endThinkTag {
			inThinkingBlock = false
			if hideThinking {
				continue
			}
		}

		if !hideThinking || !inThinkingBlock {
			// Not hiding thinking - send everything
			// Send chunk to channel if provided
			if chunkChan != nil {
				chunkChan <- text
			}
			fullResponse.WriteString(text)
		}

	}

	// Close channel if provided
	if chunkChan != nil {
		close(chunkChan)
	}

	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("error during streaming: %w", err)
	}

	return fullResponse.String(), nil
}

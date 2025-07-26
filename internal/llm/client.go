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
func (c *Client) StreamResponse(messages []openai.ChatCompletionMessageParamUnion, hideThinking bool) (string, error) {
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

	// Display streaming response
	fmt.Println("\nResponse:")

	var fullResponse strings.Builder

	inThinkingBlock := false
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

		// If hide-thinking is enabled, filter out thinking parts
		// For this implementation, we'll display all content as-is
		// A more sophisticated implementation could look for specific markers
		if hideThinking {
			if text == "start_think" {
				inThinkingBlock = true
			}

			// Process the text based on thinking block state
			if inThinkingBlock {
				// Look for end_think in this or subsequent chunks
				if text == "end_think" {
					inThinkingBlock = false
				}
				// If no end_think found, skip the entire chunk (still in thinking block)
			} else {
				// No thinking block or already exited - print everything
				fmt.Print(text)
				fullResponse.WriteString(text)
			}
		} else {
			// Not hiding thinking - print everything
			fmt.Print(text)
			fullResponse.WriteString(text)
		}

	}

	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("error during streaming: %w", err)
	}

	return fullResponse.String(), nil
}

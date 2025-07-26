package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

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

	// Time tracking
	startTime        time.Time
	endTime          time.Time
	thinkingStart    time.Time
	thinkingDuration time.Duration
	responseStart    time.Time
	responseDuration time.Duration

	mutex sync.Mutex
}

// Config holds the configuration for the LLM client
type Config struct {
	APIKey       string
	BaseURL      string
	Model        string
	Temperature  float64
	SystemPrompt string
}

// Stats holds token and timing statistics for LLM interactions
type Stats struct {
	InputTokens  int
	OutputTokens int
	ThinkingTime time.Duration
	ResponseTime time.Duration
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

	// Display time statistics
	totalTime := c.endTime.Sub(c.startTime)
	if totalTime > 0 {
		if c.thinkingDuration > 0 || c.responseDuration > 0 {
			// Show detailed breakdown when thinking is present
			fmt.Printf("Time: Thinking %v | Response %v | Total %v\n",
				c.thinkingDuration.Round(time.Millisecond),
				c.responseDuration.Round(time.Millisecond),
				(c.thinkingDuration + c.responseDuration).Round(time.Millisecond))
		} else {
			// Show simple total time when no thinking breakdown
			fmt.Printf("Time: %v\n", totalTime.Round(time.Millisecond))
		}
	}
}

// DisplayTotalUsage shows the total token usage across all interactions
func (c *Client) DisplayTotalUsage() {
	fmt.Printf("\nTotal tokens used: Input %d | Output %d | Combined %d\n",
		c.totalInputTokens, c.totalOutputTokens,
		c.totalInputTokens+c.totalOutputTokens)
}

// GetStats returns the current interaction statistics
func (c *Client) GetStats() Stats {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return Stats{
		InputTokens:  c.currentInputTokens,
		OutputTokens: c.currentOutputTokens,
		ThinkingTime: c.thinkingDuration,
		ResponseTime: c.responseDuration,
	}
}

// StreamResponse sends a message with conversation history and streams the response
// while concurrently sending chunks to the provided channel
func (c *Client) StreamResponse(messages []openai.ChatCompletionMessageParamUnion, hideThinking bool, chunkChan chan<- string) (string, error) {
	// Reset current interaction token counts and timing
	c.mutex.Lock()
	c.currentInputTokens = 0
	c.currentOutputTokens = 0
	c.startTime = time.Now()
	c.thinkingDuration = 0
	c.responseDuration = 0
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
	var responseStarted bool

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

		// Start timing the first non-empty response content
		if !responseStarted && text != "" {
			c.mutex.Lock()
			if c.responseStart.IsZero() {
				c.responseStart = time.Now()
			}
			c.mutex.Unlock()
			responseStarted = true
		}

		// Handle thinking block transitions with timing
		if !inThinkingBlock && text == startThinkTag {
			// Entering thinking block - record response duration so far
			c.mutex.Lock()
			if !c.responseStart.IsZero() {
				c.responseDuration += time.Since(c.responseStart)
				c.responseStart = time.Time{} // Reset for next response segment
			}
			c.thinkingStart = time.Now()
			c.mutex.Unlock()
			inThinkingBlock = true
		}

		if inThinkingBlock && text == endThinkTag {
			// Exiting thinking block - record thinking duration
			c.mutex.Lock()
			if !c.thinkingStart.IsZero() {
				c.thinkingDuration += time.Since(c.thinkingStart)
				c.thinkingStart = time.Time{} // Reset for next thinking segment
			}
			c.responseStart = time.Now() // Start timing response after thinking
			c.mutex.Unlock()
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

	// Record final timing when streaming completes
	c.mutex.Lock()
	c.endTime = time.Now()

	// Record final duration for active block
	if !c.thinkingStart.IsZero() {
		// Still in thinking block at end
		c.thinkingDuration += time.Since(c.thinkingStart)
	} else if !c.responseStart.IsZero() {
		// Still in response block at end
		c.responseDuration += time.Since(c.responseStart)
	}
	c.mutex.Unlock()

	// Close channel if provided
	if chunkChan != nil {
		close(chunkChan)
	}

	if err := stream.Err(); err != nil {
		return "", fmt.Errorf("error during streaming: %w", err)
	}

	return fullResponse.String(), nil
}

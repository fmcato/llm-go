package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// GetModelInfo retrieves detailed information about the specified model
func (c *Client) GetModelInfo(model string) (*openai.Model, error) {
	ctx := context.Background()
	modelInfo, err := c.client.Models.Get(ctx, model)
	if err != nil {
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}
	return modelInfo, nil
}

// DisplayModelInfo shows detailed information about the model using Ollama API
func (c *Client) DisplayModelInfo() error {
	// Convert OpenAI BaseURL to Ollama BaseURL by removing /v1 suffix if present
	ollamaBaseURL := strings.TrimSuffix(c.config.BaseURL, "/v1")

	// Use direct HTTP call with authentication
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Get model list from Ollama API
	modelsURL := ollamaBaseURL + "/api/tags"
	req, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header using the OpenAI API key
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama API error %d: %s", resp.StatusCode, string(body))
	}

	var modelsResponse struct {
		Models []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	// Find the specific model
	var modelInfo *struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}

	for _, model := range modelsResponse.Models {
		if model.Name == c.config.Model {
			modelInfo = &model
			break
		}
	}

	if modelInfo == nil {
		fmt.Printf("Model '%s' not found on the server.\n", c.config.Model)
		fmt.Println("Available models:")
		for _, model := range modelsResponse.Models {
			fmt.Printf("  - %s\n", model.Name)
		}
		return fmt.Errorf("model not found")
	}

	// Get detailed model information from /api/show
	detailsURL := ollamaBaseURL + "/api/show"
	detailsReqBody := fmt.Sprintf(`{"model":"%s"}`, c.config.Model)
	detailsReq, err := http.NewRequest("POST", detailsURL, strings.NewReader(detailsReqBody))
	if err != nil {
		return fmt.Errorf("failed to create details request: %w", err)
	}

	if c.config.APIKey != "" {
		detailsReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}
	detailsReq.Header.Set("Content-Type", "application/json")

	detailsResp, err := client.Do(detailsReq)
	var detailsResponse struct {
		Details struct {
			Family            string `json:"family"`
			ParameterSize     string `json:"parameter_size"`
			QuantizationLevel string `json:"quantization_level"`
		} `json:"details"`
		ModelInfo  map[string]interface{} `json:"model_info"`
		Template   string                 `json:"template"`
		Parameters string                 `json:"parameters"`
	}

	parameterSize := "Unknown"
	family := "Unknown"
	quantization := "Unknown"

	var allInfo map[string]interface{}
	if err == nil {
		defer detailsResp.Body.Close()
		if detailsResp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(detailsResp.Body).Decode(&detailsResponse); err == nil {
				// Extract other details
				if detailsResponse.Details.ParameterSize != "" {
					parameterSize = detailsResponse.Details.ParameterSize
				}
				if detailsResponse.Details.Family != "" {
					family = detailsResponse.Details.Family
				}
				if detailsResponse.Details.QuantizationLevel != "" {
					quantization = detailsResponse.Details.QuantizationLevel
				}
				allInfo = detailsResponse.ModelInfo
			}
		}
	}

	fmt.Println("Model Information:")
	fmt.Printf("  Name: %s\n", modelInfo.Name)
	fmt.Printf("  Size: %d MB\n", modelInfo.Size/(1024*1024))
	fmt.Printf("  Family: %s\n", family)
	fmt.Printf("  Parameters: %s\n", parameterSize)
	fmt.Printf("  Quantization: %s\n", quantization)
	fmt.Printf("  API Endpoint: %s\n", ollamaBaseURL)
	out, err := json.MarshalIndent(allInfo, "", "  ")
	fmt.Println(string(out))

	return nil
}

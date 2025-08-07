package llm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OllamaModelInfo contains detailed information about an Ollama model
type OllamaModelInfo struct {
	Name          string                 `json:"name"`
	SizeMB        int                    `json:"size_mb"`
	Family        string                 `json:"family"`
	ParameterSize string                 `json:"parameter_size"`
	Quantization  string                 `json:"quantization"`
	APIEndpoint   string                 `json:"api_endpoint"`
	Details       map[string]interface{} `json:"details"`
}

// GetOllamaModelInfo retrieves detailed information about the specified model using Ollama API
func GetOllamaModelInfo(ollamaBaseURL, apiKey, model string) (*OllamaModelInfo, error) {
	// Use direct HTTP call with authentication
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Properly format base URL without double slashes
	baseURL := strings.TrimRight(ollamaBaseURL, "/")
	modelsURL := fmt.Sprintf("%s/api/tags", baseURL)
	req, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header using the API key
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama API: %w", err)
	}
	defer resp.Body.Close()

	// Read the entire response body to handle both JSON and non-JSON responses
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("model '%s' not found on server", model)
		}
		return nil, fmt.Errorf("Ollama API error %d: %s", resp.StatusCode, string(body))
	}

	// Try parsing as JSON regardless of content type
	var modelsResponse struct {
		Models []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &modelsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w. Body: %s", err, string(body))
	}

	// Find the specific model
	var modelInfo *struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}

	for _, m := range modelsResponse.Models {
		if m.Name == model {
			modelInfo = &m
			break
		}
	}

	if modelInfo == nil {
		return nil, fmt.Errorf("model '%s' not found on the server", model)
	}

	detailsURL := fmt.Sprintf("%s/api/show", baseURL)
	detailsReqBody := fmt.Sprintf(`{"model":"%s"}`, model)
	detailsReq, err := http.NewRequest("POST", detailsURL, strings.NewReader(detailsReqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create details request: %w", err)
	}

	if apiKey != "" {
		detailsReq.Header.Set("Authorization", "Bearer "+apiKey)
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

	var detailInfo map[string]interface{}
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
				detailInfo = detailsResponse.ModelInfo
			}
		}
	}

	// Create and return the model info struct
	info := &OllamaModelInfo{
		Name:          modelInfo.Name,
		SizeMB:        int(modelInfo.Size / (1024 * 1024)),
		Family:        family,
		ParameterSize: parameterSize,
		Quantization:  quantization,
		APIEndpoint:   ollamaBaseURL,
		Details:       detailInfo,
	}

	return info, nil
}

// CheckModelExists verifies if a model exists on the Ollama server
func CheckModelExists(ollamaBaseURL, apiKey, model string) (bool, error) {
	_, err := GetOllamaModelInfo(ollamaBaseURL, apiKey, model)
	if err != nil {
		// Check for specific "not found" error
		if strings.Contains(err.Error(), "model not found") ||
			strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

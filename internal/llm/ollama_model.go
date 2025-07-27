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

	// Get model list from Ollama API
	modelsURL := ollamaBaseURL + "/api/tags"
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama API error %d: %s", resp.StatusCode, string(body))
	}

	var modelsResponse struct {
		Models []struct {
			Name string `json:"name"`
			Size int64  `json:"size"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
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

	// Get detailed model information from /api/show
	detailsURL := ollamaBaseURL + "/api/show"
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

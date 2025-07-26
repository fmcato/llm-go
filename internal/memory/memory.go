package memory

import (
	"github.com/openai/openai-go"
)

// Memory manages conversation history
type Memory struct {
	messages []openai.ChatCompletionMessageParamUnion
}

// NewMemory creates a new memory instance
func NewMemory() *Memory {
	return &Memory{
		messages: make([]openai.ChatCompletionMessageParamUnion, 0),
	}
}

// AddMessage adds a message to the conversation history
func (m *Memory) AddMessage(message openai.ChatCompletionMessageParamUnion) {
	m.messages = append(m.messages, message)
}

// AddUserMessage adds a user message to the conversation history
func (m *Memory) AddUserMessage(content string) {
	m.messages = append(m.messages, openai.UserMessage(content))
}

// AddAssistantMessage adds an assistant message to the conversation history
func (m *Memory) AddAssistantMessage(content string) {
	m.messages = append(m.messages, openai.AssistantMessage(content))
}

// AddSystemMessage adds a system message to the conversation history
func (m *Memory) AddSystemMessage(content string) {
	m.messages = append(m.messages, openai.SystemMessage(content))
}

// GetMessages returns the conversation history
func (m *Memory) GetMessages() []openai.ChatCompletionMessageParamUnion {
	return m.messages
}

// Clear clears the conversation history
func (m *Memory) Clear() {
	m.messages = make([]openai.ChatCompletionMessageParamUnion, 0)
}

// Len returns the number of messages in the conversation history
func (m *Memory) Len() int {
	return len(m.messages)
}

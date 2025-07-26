package cli

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

// CLI handles command-line interface operations
type CLI struct {
	hideThinking bool
	reader       *bufio.Reader
}

// NewCLI creates a new CLI instance
func NewCLI() *CLI {
	return &CLI{
		reader: bufio.NewReader(os.Stdin),
	}
}

// ParseFlags parses command-line flags
func (c *CLI) ParseFlags() {
	flag.BoolVar(&c.hideThinking, "hide-thinking", false, "Hide thinking/reasoning parts of the response")
	flag.Parse()
}

// GetHideThinking returns the hide-thinking flag value
func (c *CLI) GetHideThinking() bool {
	return c.hideThinking
}

// GetSystemPromptFile returns the system prompt file path from command-line arguments
func (c *CLI) GetSystemPromptFile() string {
	if flag.NArg() < 1 {
		return ""
	}
	return flag.Arg(0)
}

// ShowUsage displays usage information
func (c *CLI) ShowUsage() {
	fmt.Println("Usage: llm-go [options] <system-prompt-file>")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

// GetUserInput gets input from the user
func (c *CLI) GetUserInput() (string, error) {
	fmt.Print("\nEnter your message (or '/quit' to exit): ")
	message, err := c.reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return strings.TrimSpace(message), nil
}

// ShowResponse displays a response
func (c *CLI) ShowResponse(response string) {
	fmt.Println("\nResponse:")
	fmt.Print(response)
}

// ShowError displays an error message
func (c *CLI) ShowError(err error) {
	fmt.Printf("Error: %v\n", err)
}

// ShowGoodbye displays the goodbye message
func (c *CLI) ShowGoodbye() {
	fmt.Println("Goodbye!")
}

// ShouldQuit checks if the user wants to quit
func (c *CLI) ShouldQuit(message string) bool {
	return message == "/quit"
}

// IsValidMessage checks if the message is valid (not empty)
func (c *CLI) IsValidMessage(message string) bool {
	return message != ""
}

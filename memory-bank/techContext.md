# Technical Context

## Technologies Used
- **Language**: Go (Golang) 1.21+
- **Package Manager**: Go modules (go.mod, go.sum)
- **External Libraries**:
  - `github.com/joho/godotenv` - Environment variable loading
  - `github.com/openai/openai-go` - OpenAI API client
- **Build Tools**: Standard Go toolchain (go build, go run, go test)
- **Development Environment**: VS Code with Go extension

## Development Setup
1. **Prerequisites**:
   - Go 1.21 or higher installed
   - Git for version control
   - Text editor or IDE (VS Code recommended)

2. **Initial Setup**:
   ```bash
   git clone <repository-url>
   cd llm-go
   go mod tidy
   ```

3. **Configuration**:
   - Create `.env` file with API credentials
   - Set required environment variables
   - Configure system prompt and model selection

4. **Running the Application**:
   ```bash
   go run main.go system-prompt.txt
   # or
   go build -o llm-go
   ./llm-go system-prompt.txt
   ```

## Technical Constraints
- **Single Binary**: Application should compile to a single executable
- **Minimal Dependencies**: Keep external dependencies to a minimum
- **Cross-platform**: Should work on Linux, macOS, and Windows
- **Memory Efficiency**: Optimize for low memory usage
- **Network Resilience**: Handle network timeouts and errors gracefully
- **Configuration Validation**: Ensure all configuration values are valid

## Dependencies
### Direct Dependencies
- `github.com/joho/godotenv v1.4.0` - Loading .env files
- `github.com/openai/openai-go v0.1.0` - OpenAI API client

### Indirect Dependencies
- Various standard library packages (context, fmt, os, etc.)
- HTTP client libraries used by openai-go

## Tool Usage Patterns
- **Go Modules**: `go mod tidy` for dependency management
- **Testing**: (Future) `go test ./...` for unit and integration tests
- **Building**: `go build -o llm-go` for creating executables
- **Running**: `go run main.go system-prompt.txt` for development execution
- **Formatting**: `go fmt ./...` for code formatting
- **Vetting**: `go vet ./...` for static analysis
- **Linting**: (Future) `golangci-lint` for comprehensive linting

## Development Workflow
1. **Feature Development**:
   - Create feature branch
   - Implement changes in appropriate internal package
   - Add tests for new functionality
   - Run tests and linting
   - Commit and push

2. **Code Quality**:
   - Follow Go naming conventions
   - Use proper error handling with error wrapping
   - Write clear documentation and comments
   - Maintain consistent formatting
   - Use interfaces for testability

3. **Testing**:
   - (Future) Unit tests for all packages
   - (Future) Integration tests for API interactions
   - (Future) CLI interaction tests
   - Manual testing of CLI interactions

## Performance Considerations
- **Streaming**: Real-time response processing to reduce memory usage
- **Connection Reuse**: (Future) HTTP connection pooling for API calls
- **Memory Management**: Efficient handling of large responses
- **Concurrency**: (Future) Support for concurrent requests if needed
- **Garbage Collection**: Minimize allocations in hot paths

## Configuration Options
### Environment Variables
- `OPENAI_API_KEY` - Required API key for authentication
- `OPENAI_BASE_URL` - API endpoint URL (defaults to OpenAI)
- `OPENAI_MODEL` - Model to use (defaults to "gpt-4o")
- `OPENAI_TEMPERATURE` - Response creativity (0.0-2.0, defaults to 0.7)

### Command Line Flags
- `--hide-thinking` - Hide thinking/reasoning parts of responses
- `system-prompt-file` - Path to system prompt file (required positional argument)

## Build Configuration
### go.mod
```go
module llm-go

go 1.21

require (
    github.com/joho/godotenv v1.4.0
    github.com/openai/openai-go v0.1.0
)
```

### Build Commands
```bash
# Development build
go build -o llm-go

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o llm-go-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o llm-go-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o llm-go-windows-amd64.exe
```

## Package Structure
The application now uses Go's internal package structure for better organization:

```
llm-go/
├── main.go                 # Application entry point
├── go.mod                  # Module dependencies
├── go.sum                  # Dependency checksums
├── README.md              # Documentation
└── internal/              # Internal packages (not importable by external code)
    ├── config/            # Configuration management
    │   └── config.go      # Environment and file-based configuration
    ├── llm/               # LLM client implementation
    │   └── client.go      # OpenAI client with streaming and token tracking
    ├── memory/            # Conversation history management
    │   └── memory.go      # In-memory conversation storage
    └── cli/               # Command-line interface
        └── cli.go         # CLI handling and user interaction
```

## Development Workflow
1. **Feature Development**:
   - Create feature branch
   - Implement changes in appropriate internal package
   - Add tests for new functionality
   - Run tests and linting
   - Commit and push

2. **Code Quality**:
   - Follow Go naming conventions
   - Use proper error handling with error wrapping
   - Write clear documentation and comments
   - Maintain consistent formatting
   - Use interfaces for testability

3. **Testing**:
   - (Future) Unit tests for all packages
   - (Future) Integration tests for API interactions
   - (Future) CLI interaction tests
   - Manual testing of CLI interactions

## Security Considerations
- **API Key Management**: Never commit API keys to version control
- **Environment Variables**: Use .env files for local development
- **Input Validation**: Validate all user inputs and configuration values
- **Error Messages**: Avoid exposing sensitive information in error messages

## Future Technical Enhancements
- **Docker Support**: Containerized deployment
- **Configuration Files**: JSON/YAML configuration support
- **Logging**: Structured logging with levels
- **Metrics**: Performance and usage metrics collection
- **Caching**: Response caching for improved performance
- **Rate Limiting**: Built-in rate limiting for API calls

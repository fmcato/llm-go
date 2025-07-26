# LLM Go Client

A Go program that calls an OpenAI-compatible API in streaming mode.

## Architecture

The application follows a clean, modular architecture with clear separation of concerns:

```
main.go          # Entry point and minimal coordination
internal/
├── llm/         # LLM client interface and OpenAI implementation
├── config/      # Configuration management and validation
├── memory/      # Conversation history and persistence
└── cli/         # Command-line interface handling
```

## Features

- Uses the official OpenAI Go client library
- Configurable endpoint and API key via environment variables
- System prompt loaded from a separate file (mandatory argument)
- Interactive message input
- Streaming response display
- Optional hiding of thinking parts with a boolean flag

## Installation

```bash
go build
```

## Usage

The program automatically loads environment variables from a `.env` file if it exists in the current directory.

Set the required environment variables in the `.env` file:

```env
OPENAI_API_KEY=your-api-key-here
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_MODEL=gpt-4o  # Optional, defaults to gpt-4o
OPENAI_TEMPERATURE=0.7  # Optional, defaults to 0.7 (range 0.0-2.0)
```

Or set them manually:

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"  # Optional, defaults to OpenAI
export OPENAI_MODEL="gpt-4o"  # Optional, defaults to gpt-4o
export OPENAI_TEMPERATURE=0.7  # Optional, defaults to 0.7 (range 0.0-2.0)
```

Create a system prompt file (e.g., `system-prompt.txt`):

```
You are a helpful assistant.
```

Run the program with the system prompt file as a mandatory argument:

```bash
./llm-go system-prompt.txt
```

To hide thinking/reasoning parts of the response:

```bash
./llm-go --hide-thinking system-prompt.txt
```

## Example

Edit the `.env` file and replace `your-api-key-here` with your actual API key, create a system prompt file, then run:

```bash
./llm-go system-prompt.txt
```

Then enter your message when prompted.

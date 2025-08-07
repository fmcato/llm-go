# LLM Go Client

A Go program that calls an OpenAI-compatible API in streaming mode.

Note this is a vibe-coding exercise. The purpose is learning more about developing with LLMs using LLMs to develop (dawg).

** So use at your own risk. **

Tools:
- My private ollama server running on a shoebox
- Cline (https://cline.bot/)
- OpenRouter (https://openrouter.ai/)
  - Only free models:
    - Plan mode: deepseek-r1-0528:free
    - Act mode: qwen3-coder-480b-a35b-07-25:free

Rest of the README is LLM-generated, looks very verbose and bombastic.

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
- System prompt loaded from a separate file (optional via --system-prompt flag)
- Interactive message input
- Streaming response display
- Optional hiding of thinking parts with a boolean flag
- JSON output mode for scripting and automation
- **Model Validation**: When using the `-model` flag, the application verifies the model exists on the Ollama server before proceeding.

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

Run the program without a system prompt (uses model's default):

```bash
./llm-go
```

Run the program with a system prompt file using the explicit flag:

```bash
./llm-go --system-prompt system-prompt.txt
```

To hide thinking/reasoning parts of the response:

```bash
./llm-go --hide-thinking --system-prompt system-prompt.txt
```

## JSON Output for Scripting

The `--json` flag enables machine-readable JSON output, making it easy to integrate llm-go into scripts and automation workflows:

```bash
# Get JSON response
echo "What is 2+2?" | ./llm-go --json

# With system prompt
echo "What is 2+2?" | ./llm-go --json --system-prompt system-prompt.txt

# Hide thinking blocks in JSON output
echo "What is 2+2?" | ./llm-go --json --hide-thinking --system-prompt system-prompt.txt
```

JSON output includes the response, thinking blocks (if not hidden), and detailed statistics:

```json
{
  "response": "2 + 2 = 4.",
  "thinking": "<think>\nOkay, the user is asking \"What is 2+2?\" Let me think. The answer is straightforward. 2 plus 2 equals 4. I should just state that clearly...\n</think>",
  "stats": {
    "tokens": {
      "input": 32,
      "output": 115,
      "total": 147
    },
    "time": {
      "thinking_ms": 7446,
      "response_ms": 722,
      "total_ms": 8168
    }
  }
}
```

## Scripting Examples

### Simple question-answering script:
```bash
#!/bin/bash
question="$1"
answer=$(echo "$question" | ./llm-go --json | jq -r '.response')
echo "Q: $question"
echo "A: $answer"
```

### Batch processing with JSON output:
```bash
#!/bin/bash
for question in "What is 2+2?" "What is the capital of France?" "What is 10*5?"; do
  echo "$question" | ./llm-go --json --system-prompt system-prompt.txt
done
```

### Extract just the response text:
```bash
echo "What is 2+2?" | ./llm-go --json | jq -r '.response'
```

### Process multiple questions and save results:
```bash
echo -e "What is 2+2?\nWhat is the capital of France?" | \
while IFS= read -r question; do
  echo "$question" | ./llm-go --json --system-prompt system-prompt.txt
done > results.jsonl
```

## Example

Edit the `.env` file and replace `your-api-key-here` with your actual API key, then run:

```bash
./llm-go
```

Or with a system prompt file:

```bash
./llm-go --system-prompt system-prompt.txt
```

Then enter your message when prompted.

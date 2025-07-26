# Product Context

## Problem Statement
Users need a simple, efficient way to interact with Large Language Models from the command line. Existing solutions are often complex, heavyweight, or lack essential features like streaming responses and proper configuration management.

## Solution Overview
LLM-Go provides a lightweight, fast, and configurable CLI tool for interacting with LLMs. It focuses on simplicity and performance while maintaining flexibility for different use cases.

## User Experience Goals
- **Simplicity**: Easy to install and use with minimal setup
- **Performance**: Fast response times with streaming output
- **Flexibility**: Support for multiple LLM providers and models
- **Reliability**: Robust error handling and configuration management
- **Extensibility**: Clean architecture that allows for future enhancements

## Key Features
- Streaming responses for real-time interaction
- Environment-based configuration with .env file support
- Support for OpenAI-compatible APIs (OpenAI, Ollama, custom endpoints)
- Configurable system prompts and model selection
- Command-line flags for runtime options

## Success Metrics
- Fast response times (sub-second connection establishment)
- Low memory footprint
- High user satisfaction with CLI experience
- Minimal configuration required for basic usage
- Reliable error handling and recovery

# examples

Examples and CLI utilities for the [TAU](https://github.com/tailored-agentic-units) ecosystem.

This module sits at the top of the TAU dependency tree, importing all TAU libraries including provider sub-modules with cloud SDK dependencies. It serves as the integration point where opt-in dependencies are selected.

## Structure

```
examples/
  cmd/prompt-agent/    -- Multi-protocol CLI tool
  orchestrate/         -- Orchestrate library examples
  internal/agentutil/  -- Shared agent creation helper
  docker-compose.yml   -- Ollama with auto-model-pull
```

## Quick Start

1. Start Ollama:
   ```bash
   docker compose up -d
   ```

2. Wait for models to pull (ministral-3:8b, embeddinggemma:300m). Check progress with:
   ```bash
   docker compose logs -f
   ```

3. Test:
   ```bash
   go run ./cmd/prompt-agent \
     -config ./cmd/prompt-agent/config.ollama.json \
     -prompt "Hello"
   ```

## Sub-READMEs

- [prompt-agent](./cmd/prompt-agent/README.md) -- Multi-protocol CLI tool (chat, vision, tools, embeddings)
- [orchestrate examples](./orchestrate/README.md) -- Hub communication, state graphs, chains, parallel execution, checkpointing, conditional routing, and multi-agent workflows

## Models

| Model | Size | Capabilities | Used By |
|-------|------|-------------|---------|
| ministral-3:8b | 6.0GB | chat, vision, tools | prompt-agent, orchestrate examples |
| embeddinggemma:300m | 0.6GB | embeddings | prompt-agent embeddings |

## Cloud Providers

Azure OpenAI and AWS Bedrock configurations are available in `cmd/prompt-agent/` for testing against cloud providers. See `config.azure.json` and `config.bedrock.json` respectively. Azure infrastructure scripts are in `scripts/azure/`.

## Dependencies

This module imports the full TAU stack:

| Library | Description |
|---------|-------------|
| `github.com/tailored-agentic-units/protocol` | Core types, messages, configuration |
| `github.com/tailored-agentic-units/format` | Format registry and factory |
| `github.com/tailored-agentic-units/format/openai` | OpenAI message format |
| `github.com/tailored-agentic-units/format/converse` | AWS Converse message format |
| `github.com/tailored-agentic-units/provider` | Provider registry and factory |
| `github.com/tailored-agentic-units/provider/ollama` | Ollama provider |
| `github.com/tailored-agentic-units/provider/azure` | Azure OpenAI provider |
| `github.com/tailored-agentic-units/provider/bedrock` | AWS Bedrock provider |
| `github.com/tailored-agentic-units/agent` | Agent creation and protocol execution |
| `github.com/tailored-agentic-units/orchestrate` | Orchestration patterns (hubs, chains, graphs) |

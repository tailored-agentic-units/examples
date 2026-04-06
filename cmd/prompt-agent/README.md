# prompt-agent

Multi-protocol CLI tool for testing TAU agent configurations and LLM provider integrations.

## Usage

```
go run ./cmd/prompt-agent [flags]
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `config.json` | Path to agent configuration file |
| `-protocol` | `chat` | Protocol: `chat`, `vision`, `tools`, `embeddings` |
| `-prompt` | *(required)* | Prompt text to send |
| `-system-prompt` | | System prompt (overrides config) |
| `-auth-type` | | Provider auth type (sets `options.auth_type`) |
| `-token` | | Provider auth token (sets `options.token`) |
| `-stream` | `false` | Enable streaming responses |
| `-images` | | Comma-separated image URLs or local paths (for vision) |
| `-tools-file` | | Path to JSON file with tool definitions (for tools) |

### Configurations

| Config | Provider | Model | Protocols |
|--------|----------|-------|-----------|
| `config.ollama.json` | Ollama | ministral-3:8b | chat, vision, tools |
| `config.embedding.json` | Ollama | embeddinggemma:300m | embeddings |
| `config.azure.json` | Azure OpenAI | gpt-5-mini | chat, vision, tools |
| `config.bedrock.json` | AWS Bedrock | claude-haiku-4.5 | chat, vision, tools |

## Examples

All commands run from the `~/tau/examples` directory.

### Ollama — Ministral (chat + vision + tools)

```bash
# Chat
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.ollama.json \
  -prompt "What are the three pillars of observability?"

# Chat (streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.ollama.json \
  -prompt "What are the three pillars of observability?" \
  -stream

# Vision (remote URL)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.ollama.json \
  -prompt "Describe what you see in this image." \
  -protocol vision \
  -images "https://w.wallhaven.cc/full/zp/wallhaven-zpoxyj.png"

# Vision (remote URL, streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.ollama.json \
  -prompt "Describe what you see in this image." \
  -protocol vision \
  -images "https://w.wallhaven.cc/full/zp/wallhaven-zpoxyj.png" \
  -stream

# Vision (local file)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.ollama.json \
  -prompt "Describe what you see in this image." \
  -protocol vision \
  -images "<path-to-local-image>"

# Vision (local file, streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.ollama.json \
  -prompt "Describe what you see in this image." \
  -protocol vision \
  -images "<path-to-local-image>" \
  -stream

# Tools
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.ollama.json \
  -prompt "What is the weather in Tokyo and London?" \
  -protocol tools \
  -tools-file ./cmd/prompt-agent/tools.json
```

### Ollama — Embeddings

```bash
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.embedding.json \
  -prompt "Kubernetes is a container orchestration platform" \
  -protocol embeddings
```

### Azure — GPT-5-mini (chat + vision + tools)

Requires `AZURE_KEY` environment variable:

```bash
export AZURE_KEY="your-api-key"
```

```bash
# Chat
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.azure.json \
  -auth-type api_key -token $AZURE_KEY \
  -prompt "What are the SOLID principles?"

# Chat (streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.azure.json \
  -auth-type api_key -token $AZURE_KEY \
  -prompt "What are the SOLID principles?" \
  -stream

# Vision (remote URL)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.azure.json \
  -auth-type api_key -token $AZURE_KEY \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "https://w.wallhaven.cc/full/zp/wallhaven-zpoxyj.png"

# Vision (remote URL, streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.azure.json \
  -auth-type api_key -token $AZURE_KEY \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "https://w.wallhaven.cc/full/zp/wallhaven-zpoxyj.png" \
  -stream

# Vision (local file)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.azure.json \
  -auth-type api_key -token $AZURE_KEY \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "<path-to-local-image>"

# Vision (local file, streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.azure.json \
  -auth-type api_key -token $AZURE_KEY \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "<path-to-local-image>" \
  -stream

# Tools
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.azure.json \
  -auth-type api_key -token $AZURE_KEY \
  -prompt "What is the weather in Tokyo?" \
  -protocol tools \
  -tools-file ./cmd/prompt-agent/tools.json
```

### Bedrock — Claude Haiku (chat + vision + tools)

Requires AWS credentials configured via the default chain (`~/.aws/credentials` or environment variables).

```bash
# Chat
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "What is infrastructure as code?"

# Chat (streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "What is infrastructure as code?" \
  -stream

# Vision (remote URL)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "https://w.wallhaven.cc/full/zp/wallhaven-zpoxyj.png"

# Vision (remote URL, streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "https://w.wallhaven.cc/full/zp/wallhaven-zpoxyj.png" \
  -stream

# Vision (local file)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "<path-to-local-image>"

# Vision (local file, streaming)
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "Describe what you see." \
  -protocol vision \
  -images "<path-to-local-image>" \
  -stream

# Tools
go run ./cmd/prompt-agent \
  -config ./cmd/prompt-agent/config.bedrock.json \
  -prompt "What is the weather in Tokyo?" \
  -protocol tools \
  -tools-file ./cmd/prompt-agent/tools.json
```

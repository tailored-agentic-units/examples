package agentutil

import (
	"fmt"

	"github.com/tailored-agentic-units/agent"
	"github.com/tailored-agentic-units/format"
	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	"github.com/tailored-agentic-units/provider"
)

// NewAgent creates an agent from config with provider and format resolved from registries.
func NewAgent(cfg *config.AgentConfig) (agent.Agent, error) {
	p, err := provider.Create(cfg.Provider)
	if err != nil {
		return nil, fmt.Errorf("create provider: %w", err)
	}
	f, err := format.Create(cfg.Format)
	if err != nil {
		return nil, fmt.Errorf("create format: %w", err)
	}
	return agent.New(cfg, p, f), nil
}

// BuildMessages creates a message slice from a user prompt, prepending the
// agent's configured system prompt if one exists.
func BuildMessages(cfg *config.AgentConfig, prompt string) []protocol.Message {
	var messages []protocol.Message
	if cfg.SystemPrompt != "" {
		messages = append(messages, protocol.SystemMessage(cfg.SystemPrompt))
	}
	messages = append(messages, protocol.UserMessage(prompt))
	return messages
}

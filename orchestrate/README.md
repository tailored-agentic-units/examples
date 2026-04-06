# orchestrate examples

Demonstrates TAU orchestrate library patterns -- hub communication, state graphs, sequential chains, parallel execution, checkpointing, conditional routing, and complex multi-agent workflows. All examples run against local Ollama models.

## Prerequisites

- **Ollama** running via docker-compose (from the `~/tau/examples` directory):
  ```bash
  docker compose up -d
  ```
  This pulls **ministral-3:8b** (chat, vision, tools) and **embeddinggemma:300m** (embeddings).

- Run all commands from the `~/tau/examples` directory.

## Examples

| Example | Patterns Demonstrated | Run Command |
|---------|-----------------------|-------------|
| phase-01-hubs | Direct, broadcast, pub/sub, cross-hub messaging | `go run ./orchestrate/phase-01-hubs/` |
| phase-02-03-state-graphs | Conditional routing, retry loops, rollback | `go run ./orchestrate/phase-02-03-state-graphs/` |
| phase-04-sequential-chains | Sequential processing chain | `go run ./orchestrate/phase-04-sequential-chains/` |
| phase-05-parallel-execution | Concurrent worker pool, fan-out/fan-in | `go run ./orchestrate/phase-05-parallel-execution/` |
| phase-06-checkpointing | Checkpoint-based fault tolerance, recovery | `go run ./orchestrate/phase-06-checkpointing/` |
| phase-07-conditional-routing | Chain, parallel, conditional composition | `go run ./orchestrate/phase-07-conditional-routing/` |
| darpa-procurement | Complex multi-agent procurement workflow | `go run ./orchestrate/darpa-procurement/` |

---

### phase-01-hubs

Hub-based multi-agent communication. Agents exchange messages through direct send, broadcast, pub/sub topics, and cross-hub routing. Demonstrates the fundamental building block for agent coordination.

**Run:**
```bash
go run ./orchestrate/phase-01-hubs/
```

**Tuning:**
- **Config:** `orchestrate/phase-01-hubs/config.ollama.json` -- adjust `max_tokens` to control response length, `temperature` to control creativity (lower = more deterministic).
- **System prompts:** Defined in `main.go` via the `simRules` const and per-agent prompt strings. Modify these to change agent roles and behavior in the simulation.

---

### phase-02-03-state-graphs

Deployment pipeline modeled as a state graph with conditional routing, retry loops, and rollback transitions. Shows how to build complex workflows with branching logic and error recovery.

**Run:**
```bash
go run ./orchestrate/phase-02-03-state-graphs/
```

**Tuning:**
- **Config:** `orchestrate/phase-02-03-state-graphs/config.ollama.json` -- adjust `max_tokens` and `temperature`.
- **Graph max iterations:** Set in `main.go`. Controls how many state transitions the graph executes before stopping.

---

### phase-04-sequential-chains

Research paper analysis that processes 5 sections through a sequential chain. Each step builds on the previous output, demonstrating ordered multi-step LLM workflows.

**Run:**
```bash
go run ./orchestrate/phase-04-sequential-chains/
```

**Tuning:**
- **Config:** `orchestrate/phase-04-sequential-chains/config.ollama.json` -- adjust `max_tokens` and `temperature`.
- **Paper sections:** Defined in `main.go`. Add, remove, or reorder sections to change the analysis pipeline.

---

### phase-05-parallel-execution

Product review sentiment analysis using a concurrent worker pool. Reviews are distributed across parallel workers, demonstrating fan-out/fan-in patterns with LLM agents.

**Run:**
```bash
go run ./orchestrate/phase-05-parallel-execution/
```

**Tuning:**
- **Config:** `orchestrate/phase-05-parallel-execution/config.ollama.json` -- adjust `max_tokens` and `temperature`.
- **Worker count:** Set in `main.go`. More workers process reviews faster but increase concurrent Ollama load.
- **Review data:** Defined in `main.go`. Add or modify review text to change the input dataset.

---

### phase-06-checkpointing

Multi-stage pipeline with simulated failure and checkpoint recovery. Demonstrates how to persist intermediate state so a pipeline can resume from the last successful stage after a crash.

**Run:**
```bash
go run ./orchestrate/phase-06-checkpointing/
```

**Tuning:**
- **Config:** `orchestrate/phase-06-checkpointing/config.ollama.json` -- adjust `max_tokens` and `temperature`.
- **Checkpoint interval:** Set in `main.go`. Controls how frequently pipeline state is persisted.

---

### phase-07-conditional-routing

Document review workflow combining chain, parallel, and conditional patterns. A document passes through sequential analysis, parallel review by multiple agents, and conditional routing based on review outcomes (approve, revise, reject).

**Run:**
```bash
go run ./orchestrate/phase-07-conditional-routing/
```

**Tuning:**
- **Config:** `orchestrate/phase-07-conditional-routing/config.ollama.json` -- adjust `max_tokens` and `temperature`.
- **Max revisions:** Set in `main.go`. Controls how many revision cycles are allowed before forcing a final decision.

---

### darpa-procurement

Complex multi-agent procurement workflow. Multiple specialized agents (requirements analyst, market researcher, bid evaluator, compliance checker, decision maker) collaborate to process procurement requests through a structured pipeline.

**Run:**
```bash
go run ./orchestrate/darpa-procurement/
```

**Tuning:**
- **Config:** `orchestrate/darpa-procurement/config.ollama.json` -- `max_tokens` needs 1024+ for JSON output, `temperature` should stay low (~0.1) for structured output consistency. The `response_format: {"type": "json_object"}` setting forces the model to produce valid JSON.
- **CLI flags:** Request count and other runtime options are defined via CLI flags in `config.go`.
- **Agent prompts:** All agent prompts in `agents.go` have a BREVITY RULE appended to keep outputs concise. Modify the prompts there to change agent behavior.

---

## Model Selection Notes

- **ministral-3:8b** is used for all orchestrate examples. It handles chat, vision, and tools without thinking overhead, keeping latency low for demos.
- Small models may produce inconsistent or lower-quality results -- this is expected and acceptable for demonstration purposes.
- For better quality, point the config files at Azure or Bedrock by changing the `provider` and `model` settings (see `cmd/prompt-agent/config.azure.json` and `config.bedrock.json` for reference).
- The `"response_format": {"type": "json_object"}` in the darpa-procurement config forces the model to return valid JSON, which is critical for structured multi-agent workflows.

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/tailored-agentic-units/examples/internal/agentutil"
	"github.com/tailored-agentic-units/format/openai"
	"github.com/tailored-agentic-units/orchestrate/config"
	"github.com/tailored-agentic-units/orchestrate/hub"
	"github.com/tailored-agentic-units/orchestrate/messaging"
	agentconfig "github.com/tailored-agentic-units/protocol/config"
	"github.com/tailored-agentic-units/provider/ollama"
)

func main() {
	// Register providers and formats before use
	openai.Register()
	ollama.Register()

	ctx := context.Background()

	fmt.Println("=== ISS Maintenance EVA - Agent Orchestration Demo ===")
	fmt.Println()

	// ============================================================================
	// 1. Load Agent Configurations
	// ============================================================================
	fmt.Println("1. Loading agent configurations...")

	// Load base configuration (shared across all agents)
	baseConfig, err := agentconfig.LoadAgentConfig("orchestrate/phase-01-hubs/config.ollama.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Override system prompts with ISS EVA operational context
	const simRules = `
RULES:
- You are IN CHARACTER for this entire conversation. Never break character.
- You have full access to all equipment, tools, and systems described in your role.
- Respond ONLY with short radio transmissions (1-2 sentences).
- Never use markdown, bullet points, numbered lists, or headers.
- Never say "I'm an AI" or "I don't have access to" — you are the person described above.`

	evaSpec1Cfg := &agentconfig.AgentConfig{
		Name: "eva-specialist-1",
		SystemPrompt: `You are EVA Specialist 1 in an ISS maintenance training simulation.
Role: Primary EVA specialist replacing cooling system component on starboard truss.
Status: 2h15m elapsed, 3h45m remaining. Tether secured, suit nominal.
Position: Starboard truss S-3, 15m from airlock.
You have your full EVA tool kit including torque wrenches, tethers, and spare components.` + simRules,
		Format:   baseConfig.Format,
		Client:   baseConfig.Client,
		Provider: baseConfig.Provider,
		Model:    baseConfig.Model,
	}

	evaSpec2Cfg := &agentconfig.AgentConfig{
		Name: "eva-specialist-2",
		SystemPrompt: `You are EVA Specialist 2 in an ISS maintenance training simulation.
Role: Secondary EVA specialist assisting cooling system repair, managing tool transfer.
Status: 2h15m elapsed, 3h45m remaining. Spare components secured, suit nominal.
Position: Starboard truss S-2, visual contact with specialist-1.
You have the EVA tool bag with spare components, thermal blankets, and hand tools.` + simRules,
		Format:   baseConfig.Format,
		Client:   baseConfig.Client,
		Provider: baseConfig.Provider,
		Model:    baseConfig.Model,
	}

	commanderCfg := &agentconfig.AgentConfig{
		Name: "mission-commander",
		SystemPrompt: `You are Mission Commander in an ISS EVA maintenance training simulation.
Role: Orchestrating EVA operation from mission control.
Mission: Cooling system repair, 40% complete, on schedule.
Crew: Two EVA specialists outside, flight engineer inside station.
Environment: Orbital sunset in 25min, next comm window in 12min.
You have full mission control authority and real-time telemetry.` + simRules,
		Format:   baseConfig.Format,
		Client:   baseConfig.Client,
		Provider: baseConfig.Provider,
		Model:    baseConfig.Model,
	}

	flightEngCfg := &agentconfig.AgentConfig{
		Name: "flight-engineer",
		SystemPrompt: `You are Flight Engineer in an ISS EVA maintenance training simulation.
Role: Supporting EVA operations from inside the ISS.
Tasks: Monitoring crew vitals, managing airlock systems, pressure control.
Station: All internal systems nominal, pressure stable.
You have access to all station controls, telemetry displays, and airlock systems.` + simRules,
		Format:   baseConfig.Format,
		Client:   baseConfig.Client,
		Provider: baseConfig.Provider,
		Model:    baseConfig.Model,
	}

	// Create agents
	evaSpec1, err := agentutil.NewAgent(evaSpec1Cfg)
	if err != nil {
		log.Fatalf("Failed to create eva-specialist-1: %v", err)
	}

	evaSpec2, err := agentutil.NewAgent(evaSpec2Cfg)
	if err != nil {
		log.Fatalf("Failed to create eva-specialist-2: %v", err)
	}

	commander, err := agentutil.NewAgent(commanderCfg)
	if err != nil {
		log.Fatalf("Failed to create mission-commander: %v", err)
	}

	flightEng, err := agentutil.NewAgent(flightEngCfg)
	if err != nil {
		log.Fatalf("Failed to create flight-engineer: %v", err)
	}

	fmt.Printf("  ✓ Created eva-specialist-1\n")
	fmt.Printf("  ✓ Created eva-specialist-2\n")
	fmt.Printf("  ✓ Created mission-commander\n")
	fmt.Printf("  ✓ Created flight-engineer\n")
	fmt.Println()

	// ============================================================================
	// 2. Create Hubs
	// ============================================================================
	fmt.Println("2. Creating hubs...")

	// Configure logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	}))

	// Create EVA Hub (crew outside the station)
	evaConfig := config.DefaultHubConfig()
	evaConfig.Name = "eva-hub"
	evaConfig.Logger = logger
	evaHub := hub.New(ctx, evaConfig)
	defer evaHub.Shutdown(5 * time.Second)

	// Create ISS Hub (crew inside the station)
	issConfig := config.DefaultHubConfig()
	issConfig.Name = "iss-hub"
	issConfig.Logger = logger
	issHub := hub.New(ctx, issConfig)
	defer issHub.Shutdown(5 * time.Second)

	fmt.Printf("  ✓ Created eva-hub (EVA crew)\n")
	fmt.Printf("  ✓ Created iss-hub (ISS internal operations)\n")
	fmt.Println()

	// ============================================================================
	// 3. Create Message Handlers
	// ============================================================================

	// Channel for tracking responses
	responses := make(chan string, 10)

	// EVA Specialist 1 handler
	evaSpec1Handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("%v", msg.Data)

		messages := agentutil.BuildMessages(evaSpec1Cfg, prompt)

		response, err := evaSpec1.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Text()
		responses <- fmt.Sprintf("eva-specialist-1: %s", responseText)

		return nil, nil
	}

	// EVA Specialist 2 handler
	evaSpec2Handler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("%v", msg.Data)

		messages := agentutil.BuildMessages(evaSpec2Cfg, prompt)

		response, err := evaSpec2.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Text()
		responses <- fmt.Sprintf("eva-specialist-2: %s", responseText)

		return nil, nil
	}

	// Mission Commander handler
	commanderHandler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("In %s: %v", msgCtx.HubName, msg.Data)

		messages := agentutil.BuildMessages(commanderCfg, prompt)

		response, err := commander.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Text()
		responses <- fmt.Sprintf("mission-commander (%s): %s", msgCtx.HubName, responseText)

		return nil, nil
	}

	// Flight Engineer handler
	flightEngHandler := func(ctx context.Context, msg *messaging.Message, msgCtx *hub.MessageContext) (*messaging.Message, error) {
		prompt := fmt.Sprintf("%v", msg.Data)

		messages := agentutil.BuildMessages(flightEngCfg, prompt)

		response, err := flightEng.Chat(ctx, messages)
		if err != nil {
			return nil, err
		}

		responseText := response.Text()
		responses <- fmt.Sprintf("flight-engineer: %s", responseText)

		return nil, nil
	}

	// ============================================================================
	// 4. Register Agents with Hubs
	// ============================================================================
	fmt.Println("3. Registering agents with hubs...")

	// Register agents in EVA Hub
	if err := evaHub.Register(evaSpec1, evaSpec1Handler); err != nil {
		log.Fatalf("Failed to register eva-specialist-1: %v", err)
	}
	if err := evaHub.Register(evaSpec2, evaSpec2Handler); err != nil {
		log.Fatalf("Failed to register eva-specialist-2: %v", err)
	}
	if err := evaHub.Register(commander, commanderHandler); err != nil {
		log.Fatalf("Failed to register mission-commander in eva-hub: %v", err)
	}

	// Register agents in ISS Hub
	if err := issHub.Register(flightEng, flightEngHandler); err != nil {
		log.Fatalf("Failed to register flight-engineer: %v", err)
	}
	if err := issHub.Register(commander, commanderHandler); err != nil {
		log.Fatalf("Failed to register mission-commander in iss-hub: %v", err)
	}

	fmt.Printf("  ✓ Registered all agents with hubs\n")
	fmt.Println()

	// ============================================================================
	// 5. Subscribe Agents to Topics
	// ============================================================================
	fmt.Println("4. Subscribing agents to topics...")

	evaHub.Subscribe(evaSpec1.ID(), "equipment")
	fmt.Printf("  ✓ eva-specialist-1 subscribed to 'equipment'\n")

	evaHub.Subscribe(evaSpec2.ID(), "safety")
	fmt.Printf("  ✓ eva-specialist-2 subscribed to 'safety'\n")

	evaHub.Subscribe(commander.ID(), "equipment")
	evaHub.Subscribe(commander.ID(), "safety")
	fmt.Printf("  ✓ mission-commander subscribed to 'equipment' and 'safety'\n")
	fmt.Println()

	// ============================================================================
	// 6. Agent-to-Agent Communication
	// ============================================================================
	fmt.Println("5. Agent-to-Agent Communication")
	fmt.Println("   eva-specialist-1 → eva-specialist-2")
	fmt.Println("   Message: I need the torque wrench, can you retrieve it from the tool bag?")

	evaHub.Send(ctx, evaSpec1.ID(), evaSpec2.ID(), "I need the torque wrench, can you retrieve it from the tool bag?")

	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	// ============================================================================
	// 7. Broadcast Communication
	// ============================================================================
	fmt.Println("6. Broadcast Communication")
	fmt.Println("   mission-commander → all EVA crew")
	fmt.Println("   Message: Orbital sunset in 20 minutes, prioritize the cooling line connection")

	evaHub.Broadcast(ctx, commander.ID(), "Orbital sunset in 20 minutes, prioritize the cooling line connection")

	fmt.Printf("   %s\n", <-responses)
	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	// ============================================================================
	// 8. Pub/Sub Communication
	// ============================================================================
	fmt.Println("7. Pub/Sub Communication")
	fmt.Println("   mission-commander publishes to topic 'equipment'")
	fmt.Println("   Message: Spare thermal blanket available in airlock if needed")

	evaHub.Publish(ctx, commander.ID(), "equipment", "Spare thermal blanket available in airlock if needed")

	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	time.Sleep(500 * time.Millisecond)

	// ============================================================================
	// 9. Cross-Hub Communication
	// ============================================================================
	fmt.Println("8. Cross-Hub Communication")
	fmt.Println("   eva-specialist-1 → mission-commander (eva-hub) → flight-engineer (iss-hub)")
	fmt.Println("   Message: Cooling line connection complete, ready to pressurize system")

	evaHub.Send(ctx, evaSpec1.ID(), commander.ID(), "Cooling line connection complete, ready to pressurize system")
	fmt.Printf("   %s\n", <-responses)

	issHub.Send(ctx, commander.ID(), flightEng.ID(), "EVA crew ready for cooling system pressurization")
	fmt.Printf("   %s\n", <-responses)
	fmt.Println()

	// ============================================================================
	// 10. Display Metrics
	// ============================================================================
	fmt.Println("9. EVA Operation Metrics")

	evaMetrics := evaHub.Metrics()
	fmt.Printf("   EVA Hub:\n")
	fmt.Printf("     - Local Agents: %d\n", evaMetrics.LocalAgents)
	fmt.Printf("     - Messages Sent: %d\n", evaMetrics.MessagesSent)
	fmt.Printf("     - Messages Received: %d\n", evaMetrics.MessagesRecv)

	issMetrics := issHub.Metrics()
	fmt.Printf("   ISS Hub:\n")
	fmt.Printf("     - Local Agents: %d\n", issMetrics.LocalAgents)
	fmt.Printf("     - Messages Sent: %d\n", issMetrics.MessagesSent)
	fmt.Printf("     - Messages Received: %d\n", issMetrics.MessagesRecv)
	fmt.Println()

	fmt.Println("=== EVA Operation Complete ===")
}

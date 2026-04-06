package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/tailored-agentic-units/agent"
	"github.com/tailored-agentic-units/format"
	"github.com/tailored-agentic-units/format/converse"
	"github.com/tailored-agentic-units/format/openai"
	"github.com/tailored-agentic-units/protocol"
	"github.com/tailored-agentic-units/protocol/config"
	"github.com/tailored-agentic-units/provider"
	"github.com/tailored-agentic-units/provider/azure"
	"github.com/tailored-agentic-units/provider/bedrock"
	"github.com/tailored-agentic-units/provider/ollama"
)

func main() {
	// Register providers and formats before use
	openai.Register()
	converse.Register()
	ollama.Register()
	azure.Register()
	bedrock.Register()

	var (
		configFile   = flag.String("config", "config.json", "Configuration file to use")
		proto        = flag.String("protocol", "chat", "Protocol to use (chat, vision, tools, embeddings)")
		prompt       = flag.String("prompt", "", "Prompt to send to the agent")
		systemPrompt = flag.String("system-prompt", "", "System prompt (overrides config)")
		token        = flag.String("token", "", "Provider auth token (sets options.token)")
		authType     = flag.String("auth-type", "", "Provider auth type (sets options.auth_type, e.g., api_key, bearer, default)")
		stream       = flag.Bool("stream", false, "Enable streaming responses")

		images    = flag.String("images", "", "Comma-separated image URLs/paths (for vision)")
		toolsFile = flag.String("tools-file", "", "JSON file containing tool definitions (for tools)")
	)
	flag.Parse()

	if *prompt == "" {
		log.Fatal("Error: -prompt flag is required")
	}

	cfg, err := config.LoadAgentConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *authType != "" || *token != "" {
		if cfg.Provider.Options == nil {
			cfg.Provider.Options = make(map[string]any)
		}
		if *authType != "" {
			cfg.Provider.Options["auth_type"] = *authType
		}
		if *token != "" {
			cfg.Provider.Options["token"] = *token
		}
	}

	if *systemPrompt != "" {
		cfg.SystemPrompt = *systemPrompt
	}

	// Create provider and format, then agent
	p, err := provider.Create(cfg.Provider)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	f, err := format.Create(cfg.Format)
	if err != nil {
		log.Fatalf("Failed to create format: %v", err)
	}

	a := agent.New(cfg, p, f)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.TimeoutDuration())
	defer cancel()

	// Build messages from prompt and optional system prompt
	var messages []protocol.Message
	if cfg.SystemPrompt != "" {
		messages = append(messages, protocol.SystemMessage(cfg.SystemPrompt))
	}
	messages = append(messages, protocol.UserMessage(*prompt))

	switch *proto {
	case "chat":
		if *stream {
			executeChatStream(ctx, a, messages)
		} else {
			executeChat(ctx, a, messages)
		}
	case "vision":
		if *images == "" {
			log.Fatal("Error: -images flag is required for vision protocol")
		}
		imageList := strings.Split(*images, ",")
		for i, img := range imageList {
			imageList[i] = strings.TrimSpace(img)
		}
		preparedImages := prepareImages(imageList)
		if *stream {
			executeVisionStream(ctx, a, messages, preparedImages)
		} else {
			executeVision(ctx, a, messages, preparedImages)
		}
	case "tools":
		if *toolsFile == "" {
			log.Fatal("Error: -tools-file flag is required for tools protocol")
		}
		toolList := loadTools(*toolsFile)
		executeTools(ctx, a, messages, toolList)
	case "embeddings":
		executeEmbeddings(ctx, a, *prompt)
	default:
		log.Fatalf("Unknown protocol: %s", *proto)
	}
}

func executeChat(
	ctx context.Context,
	a agent.Agent,
	messages []protocol.Message,
) {
	resp, err := a.Chat(ctx, messages)
	if err != nil {
		log.Fatalf("Chat failed: %v", err)
	}
	fmt.Printf("Response: %s\n", resp.Text())
	if resp.Usage != nil {
		fmt.Printf(
			"Tokens: %d prompt + %d completions = %d total",
			resp.Usage.InputTokens,
			resp.Usage.OutputTokens,
			resp.Usage.TotalTokens,
		)
	}
}

func executeChatStream(
	ctx context.Context,
	a agent.Agent,
	messages []protocol.Message,
) {
	stream, err := a.ChatStream(ctx, messages)
	if err != nil {
		log.Fatalf("ChatStream failed: %v", err)
	}

	for chunk := range stream {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}
		fmt.Print(chunk.Text())
	}
	fmt.Println()
}

func executeVision(
	ctx context.Context,
	a agent.Agent,
	messages []protocol.Message,
	images []format.Image,
) {
	resp, err := a.Vision(ctx, messages, images)
	if err != nil {
		log.Fatalf("Vision failed: %v", err)
	}
	fmt.Printf("Vision response: %s\n", resp.Text())
	if resp.Usage != nil {
		fmt.Printf(
			"Tokens: %d prompt + %d completion = %d total\n",
			resp.Usage.InputTokens,
			resp.Usage.OutputTokens,
			resp.Usage.TotalTokens,
		)
	}
}

func executeVisionStream(
	ctx context.Context,
	a agent.Agent,
	messages []protocol.Message,
	images []format.Image,
) {
	stream, err := a.VisionStream(ctx, messages, images)
	if err != nil {
		log.Fatalf("VisionStream failed: %v", err)
	}

	for chunk := range stream {
		if chunk.Error != nil {
			log.Fatalf("Stream error: %v", chunk.Error)
		}

		fmt.Print(chunk.Text())
	}

	fmt.Println()
}

func executeTools(
	ctx context.Context,
	a agent.Agent,
	messages []protocol.Message,
	tools []format.ToolDefinition,
) {
	resp, err := a.Tools(ctx, messages, tools)
	if err != nil {
		log.Fatalf("Tools failed: %v", err)
	}

	if text := resp.Text(); text != "" {
		fmt.Printf("Response: %s\n", text)
	}

	toolCalls := resp.ToolCalls()
	if len(toolCalls) > 0 {
		fmt.Printf("\nTool Calls:\n")
		for _, tc := range toolCalls {
			argsJSON, _ := json.Marshal(tc.Input)
			fmt.Printf("  - %s(%s)\n", tc.Name, string(argsJSON))
		}
	}

	if resp.Usage != nil {
		fmt.Printf("\nTokens: %d prompt + %d completion = %d total\n",
			resp.Usage.InputTokens,
			resp.Usage.OutputTokens,
			resp.Usage.TotalTokens,
		)
	}
}

func executeEmbeddings(
	ctx context.Context,
	a agent.Agent,
	input string,
) {
	resp, err := a.Embed(ctx, input)
	if err != nil {
		log.Fatalf("Embeddings failed: %v", err)
	}

	fmt.Printf("Input: %q\n\n", input)
	fmt.Printf("Generated %d embedding(s):\n\n", len(resp.Embeddings))

	for i, embedding := range resp.Embeddings {
		fmt.Printf("Embedding [%d]:\n", i)
		fmt.Printf("  Dimensions: %d\n", len(embedding))

		if len(embedding) > 0 {
			previewCount := 5
			if len(embedding) <= previewCount*2 {
				fmt.Printf("  Values: [")
				for j, val := range embedding {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%.6f", val)
				}
				fmt.Printf("]\n")
			} else {
				fmt.Printf("  Values: [")
				for j := range previewCount {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%.6f", embedding[j])
				}
				fmt.Printf(", ..., ")
				start := len(embedding) - previewCount
				for j := start; j < len(embedding); j++ {
					if j > start {
						fmt.Printf(", ")
					}
					fmt.Printf("%.6f", embedding[j])
				}
				fmt.Printf("]\n")
			}

			var sum, min, max float64
			min = embedding[0]
			max = embedding[0]

			for _, val := range embedding {
				sum += val
				if val < min {
					min = val
				}
				if val > max {
					max = val
				}
			}

			mean := sum / float64(len(embedding))
			fmt.Printf("  Statistics: min=%.6f, max=%.6f, mean=%.6f\n", min, max, mean)
		}

		fmt.Println()
	}

	if resp.Usage != nil {
		fmt.Printf("Token Usage: %d total\n", resp.Usage.TotalTokens)
	}
}

func loadTools(filename string) []format.ToolDefinition {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read tools file: %v", err)
	}

	var tools []format.ToolDefinition
	if err := json.Unmarshal(data, &tools); err != nil {
		log.Fatalf("Failed to parse tools file: %v", err)
	}

	return tools
}

func prepareImages(imageList []string) []format.Image {
	prepared := make([]format.Image, len(imageList))
	for i, img := range imageList {
		if strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") {
			// Download and encode remote images (some providers only support base64)
			data, err := downloadImage(img)
			if err != nil {
				log.Fatalf("Failed to download image %s: %v", img, err)
			}

			// Detect MIME type from downloaded content
			mimeType := http.DetectContentType(data)

			// Validate it's an image
			if !strings.HasPrefix(mimeType, "image/") {
				log.Fatalf("URL %s is not an image (detected type: %s)", img, mimeType)
			}

			prepared[i] = format.Image{
				Data:   data,
				Format: strings.TrimPrefix(mimeType, "image/"),
			}
		} else {
			// Expand home directory if needed
			if strings.HasPrefix(img, "~/") {
				home, err := os.UserHomeDir()
				if err != nil {
					log.Fatalf("Failed to get home directory: %v", err)
				}
				img = strings.Replace(img, "~", home, 1)
			}

			// Local file, read and encode
			data, err := os.ReadFile(img)
			if err != nil {
				log.Fatalf("Failed to read image %s: %v", img, err)
			}

			// Detect MIME type from content
			mimeType := http.DetectContentType(data)

			// Validate it's an image
			if !strings.HasPrefix(mimeType, "image/") {
				log.Fatalf("File %s is not an image (detected type: %s)", img, mimeType)
			}

			prepared[i] = format.Image{
				Data:   data,
				Format: strings.TrimPrefix(mimeType, "image/"),
			}
		}
	}
	return prepared
}

func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

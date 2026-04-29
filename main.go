package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

type sReq struct {
	Query string `json:"query"`
}

var (
	systemPrompt string
	llmClient    llms.Model
	temperature  float64
)

func initEnv() {
	promptData, err := os.ReadFile("prompts/system.xml")
	if err != nil {
		log.Fatalf("Failed to read system prompt: %v", err)
	}
	systemPrompt = string(promptData)

	provider := os.Getenv("AI_PROVIDER")
	if provider == "" {
		provider = "google"
	}

	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash-lite-preview-02-05"
	}

	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		log.Fatal("AI_API_KEY environment variable is required")
	}

	tempStr := os.Getenv("AI_TEMPERATURE")
	if tempStr != "" {
		temp, err := strconv.ParseFloat(tempStr, 64)
		if err == nil {
			temperature = temp
		} else {
			temperature = 0.1
		}
	} else {
		temperature = 0.1
	}

	ctx := context.Background()

	switch provider {
	case "google":
		client, err := googleai.New(
			ctx,
			googleai.WithAPIKey(apiKey),
			googleai.WithDefaultModel(model),
		)
		if err != nil {
			log.Fatalf("Failed to initialize Google AI client: %v", err)
		}
		llmClient = client

	case "openai":
		opts := []openai.Option{
			openai.WithModel(model),
			openai.WithToken(apiKey),
		}

		baseURL := os.Getenv("AI_BASE_URL")
		if baseURL != "" {
			opts = append(opts, openai.WithBaseURL(baseURL))
		}

		client, err := openai.New(opts...)
		if err != nil {
			log.Fatalf("Failed to initialize OpenAI client: %v", err)
		}
		llmClient = client

	default:
		log.Fatalf("Unsupported AI_PROVIDER: %s. Must be 'google' or 'openai'", provider)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req sReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, req.Query),
	}

	resp, err := llmClient.GenerateContent(ctx, content, llms.WithTemperature(temperature))
	if err != nil {
		log.Printf("Error generating content: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(resp.Choices) == 0 {
		http.Error(w, "No response generated", http.StatusInternalServerError)
		return
	}

	resultText := resp.Choices[0].Content

	// Try to return raw JSON response directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, resultText)
}

func main() {
	initEnv()

	http.HandleFunc("/sReq", handleSearch)

	port := "8080"
	log.Printf("Starting CineSearch Pro service on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
)

type sReq struct {
	Query string `json:"query"`
}

type Guess struct {
	Name         string `json:"name"`
	Date         string `json:"date"`
	Genre        string `json:"genre"`
	Info         string `json:"info"`
	ImdbID       string `json:"imdb_id"`
	FilmCoverArt string `json:"film_cover_art"`
}

type LLMResponse struct {
	Guesses []Guess `json:"guesses"`
}

// TMDB API structures
type TMDBFindResponse struct {
	MovieResults []TMDBResult `json:"movie_results"`
	TvResults    []TMDBResult `json:"tv_results"`
}

type TMDBResult struct {
	PosterPath string `json:"poster_path"`
}

var (
	systemPrompt string
	llmClient    llms.Model
	temperature  float64
	tmdbAPIKey   string
)

const fallbackCoverArt = "https://via.placeholder.com/500x750.png?text=Poster+Not+Found"
const tmdbImageBaseURL = "https://image.tmdb.org/t/p/w500"

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

	tmdbAPIKey = os.Getenv("TMDB_API_KEY")
	if tmdbAPIKey == "" {
		log.Println("Warning: TMDB_API_KEY is not set. Fallback posters will be used.")
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

func cleanJSONResponse(raw string) string {
	cleaned := strings.TrimSpace(raw)
	// Remove markdown block if present
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	return strings.TrimSpace(cleaned)
}

func fetchTMDBPoster(imdbID string) string {
	if tmdbAPIKey == "" || imdbID == "" {
		return fallbackCoverArt
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/find/%s?api_key=%s&external_source=imdb_id", imdbID, tmdbAPIKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		return fallbackCoverArt
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fallbackCoverArt
	}

	var tmdbResp TMDBFindResponse
	if err := json.Unmarshal(body, &tmdbResp); err != nil {
		return fallbackCoverArt
	}

	var posterPath string
	if len(tmdbResp.MovieResults) > 0 && tmdbResp.MovieResults[0].PosterPath != "" {
		posterPath = tmdbResp.MovieResults[0].PosterPath
	} else if len(tmdbResp.TvResults) > 0 && tmdbResp.TvResults[0].PosterPath != "" {
		posterPath = tmdbResp.TvResults[0].PosterPath
	}

	if posterPath != "" {
		return tmdbImageBaseURL + posterPath
	}

	return fallbackCoverArt
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

	resultText := cleanJSONResponse(resp.Choices[0].Content)

	var llmResp LLMResponse
	if err := json.Unmarshal([]byte(resultText), &llmResp); err != nil {
		log.Printf("Error parsing JSON from LLM: %v\nRaw output: %s", err, resultText)
		http.Error(w, "Failed to parse generation response", http.StatusInternalServerError)
		return
	}

	// Enrich with TMDB Posters
	for i, guess := range llmResp.Guesses {
		// Only try to fetch if we have an IMDB ID and it's not the "SYSTEM ERROR" out-of-bounds message
		if guess.ImdbID != "" && guess.ImdbID != "null" && !strings.HasPrefix(guess.Info, "SYSTEM ERROR") {
			llmResp.Guesses[i].FilmCoverArt = fetchTMDBPoster(guess.ImdbID)
		} else {
			llmResp.Guesses[i].FilmCoverArt = fallbackCoverArt
		}
	}

	finalResp, err := json.Marshal(llmResp)
	if err != nil {
		log.Printf("Error marshaling final response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(finalResp)
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

# CineSearch Pro

CineSearch Pro is a robust, production-ready microservice built in Go that identifies movies and television series based on natural language queries. It leverages a powerful dual-engine approach to return structured, deterministic JSON data for seamless backend integration.

## Architecture & Features

- **Dual-Engine Pipeline:** Utilizes an LLM engine for highly accurate natural language deduction and metadata extraction, seamlessly paired with the TMDB API to verify the `imdb_id` and fetch official, high-quality cover art.
- **Model Agnostic:** Built on top of `langchaingo`, allowing you to effortlessly switch between Google and OpenAI-compatible models.
- **Universal Provider Support:** Not limited to just OpenAI. By configuring the `AI_BASE_URL`, you can connect to local models (e.g., Ollama) or other compatible API providers (e.g., Groq, Together.ai) using standard OpenAI tooling.
- **Externalized Prompting:** The core instruction set and boundaries are maintained in a clean XML format at `prompts/system.xml`, making it easy to tweak the service's persona without touching Go code.
- **Robust Parsing:** Built-in safeguards automatically strip Markdown wrappers and sanitize outputs to guarantee standard JSON structures.
- **Dockerized:** Includes a multi-stage Dockerfile for lightweight, secure, and rapid deployments without requiring local build tools.

## Prerequisites

- Go 1.22+
- Docker and Docker Compose (optional, for containerized deployment)

## Setup and Configuration

1. **Clone the repository:**
   ```bash
   git clone <repository_url>
   cd cinesearch-pro
   ```

2. **Environment Configuration:**
   Copy the example environment file and fill in your details:
   ```bash
   cp .env.example .env
   ```

   ### Configuration Options

   | Variable | Description | Default |
   |----------|-------------|---------|
   | `AI_PROVIDER` | The LLM provider to use (`google` or `openai`). | `google` |
   | `AI_MODEL` | The specific model identifier. | `gemini-2.0-flash-lite-preview-02-05` |
   | `AI_API_KEY` | Your authentication key for the chosen provider. | **Required** |
   | `AI_BASE_URL` | Optional endpoint override for OpenAI-compatible services. | (Empty) |
   | `AI_TEMPERATURE` | Generation temperature. Lower values produce more deterministic JSON. | `0.1` |
   | `TMDB_API_KEY` | API Key for The Movie Database (used for cover art). | **Required** |

## Running the Service

### Using Docker Compose (Recommended)

To build and start the service in detached mode:

```bash
docker-compose up -d --build
```

### Running Locally

To run the application directly with Go:

```bash
go mod download
export $(cat .env | xargs) && go run main.go
```

The service will start and listen on `http://localhost:8080`.

## API Usage

The service exposes a single POST endpoint.

### `POST /sReq`

**Request:**
```json
{
  "query": "A futuristic movie where humans are plugged into a simulation, and the main character learns kung fu."
}
```

**Response (JSON):**
```json
{
  "guesses": [
    {
      "name": "The Matrix",
      "date": "1999",
      "genre": "Sci-Fi",
      "info": "A computer hacker learns from mysterious rebels about the true nature of his reality...",
      "imdb_id": "tt0133093",
      "film_cover_art": "https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg"
    },
    ...
  ]
}
```

*(Note: The system is constrained to return exactly 5 guesses based on the criteria in `prompts/system.xml`. If the TMDB lookup fails, a fallback image placeholder is provided.)*

## Contributors

[Your Name/Organization]

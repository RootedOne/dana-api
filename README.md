# CineSearch Pro

CineSearch Pro is a robust, production-ready microservice built in Go that identifies movies and television series based on natural language queries. It leverages modern LLM capabilities to return structured, deterministic JSON data for seamless backend integration.

## Architecture & Features

- **Model Agnostic:** Built on top of `langchaingo`, allowing you to seamlessly switch between Google (Gemini) and OpenAI-compatible models.
- **Universal Provider Support:** Not limited to just OpenAI. By configuring the `AI_BASE_URL`, you can connect to local models (e.g., Ollama) or other compatible API providers (e.g., Groq, Together.ai) using standard OpenAI tooling.
- **Externalized Prompting:** The core instruction set and boundaries are maintained in a clean XML format at `prompts/system.xml`, making it easy to tweak the service's persona without touching Go code.
- **Dockerized:** Includes a multi-stage Dockerfile for lightweight, secure, and rapid deployments.

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
      "film_cover_art": "..."
    },
    ...
  ]
}
```

*(Note: The system is constrained to return exactly 5 guesses based on the criteria in `prompts/system.xml`.)*


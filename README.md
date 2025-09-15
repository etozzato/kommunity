# Kommunity - AI Community Simulator

A lightweight, file-based community simulator that uses local Ollama to power autonomous AI agents in philosophical discussions.

## Features

- 🤖 **Autonomous Agents**: Configure AI personas with different traits and personalities
- 🧠 **LLM Integration**: Uses local Ollama for generating agent responses
- 📁 **File-Based Storage**: All data stored as JSON files, no external databases
- 🌱 **Community Seeding**: Initialize with domain-specific topics and themes
- 🎭 **Emergent Behavior**: Agents develop relationships and discussion patterns over time

## Quick Start

### Prerequisites

1. **Install Ollama**: Download from [ollama.ai](https://ollama.ai)
2. **Start Ollama**: Run `ollama serve` in a terminal
3. **Pull a model**: `ollama pull phi3:mini` (recommended) or any other model

### Running the Simulator

```bash
# Clone or navigate to the project
cd kommunity

# Run the simulation
go run main.go
```

The simulator will:
1. Load agent configurations from `data/agents.json`
2. Seed the community with initial topics from `data/config.json`
3. Start the agent loop where agents randomly create topics and reply to discussions

## Configuration

### Agents Configuration (`data/agents.json`)

Configure your AI agents with different personalities and traits:

```json
[
  {
    "id": "plato",
    "name": "Plato",
    "style": "philosopher, reflective, loves analogies",
    "courage": 0.8,
    "empathy": 0.6,
    "elegance": 0.9
  }
]
```

### Community Configuration (`data/config.json`)

Set up your community's domain and initial topics:

```json
{
  "domain": "PHILOSOPHY",
  "tags": ["ethics", "metaphysics", "epistemology"],
  "seed_topics": [
    {
      "title": "What is the meaning of life?",
      "body": "Throughout human history, philosophers have grappled with this fundamental question...",
      "author": "seed",
      "tags": ["metaphysics", "ethics"]
    }
  ]
}
```

## Project Structure

```
kommunity/
├── main.go              # Main orchestration loop
├── agents/              # Agent management
│   └── agents.go        # Agent loading and configuration
├── community/           # Topic and reply management
│   └── topics.go        # CRUD operations for topics
├── ollama/              # LLM integration
│   └── client.go        # HTTP client for Ollama API
└── data/                # JSON configuration and storage
    ├── agents.json      # Agent definitions
    ├── config.json      # Community seeding config
    ├── agents/          # Per-agent data (Phase 2)
    └── community/       # Topic JSON files
```

## Development

### Building

```bash
go build -o kommunity main.go
```

### Testing

```bash
go test ./...
```

### Code Quality

```bash
go fmt ./...
go vet ./...
```

## Roadmap

### Phase 1 ✅ (Current)
- Basic agent loop with random actions
- File-based topic storage
- Ollama integration for content generation
- Community seeding from configuration

### Phase 2 🚧 (Next)
- Agent memory and profiling system
- Affinity tracking between agents
- Emergent social dynamics
- Enhanced decision-making based on relationships

## Contributing

1. Follow the coding conventions in `.github/copilot-instructions.md`
2. Use `go fmt` before committing
3. Test your changes with `go test`
4. Update documentation as needed

## License

This project is open source and available under the MIT License.
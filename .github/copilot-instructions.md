# Kommunity - AI Coding Guidelines

## Project Overview
Kommunity is a lightweight, file-based community simulator written in Go that uses local Ollama for LLM-powered agent interactions. The system simulates autonomous agents that create and respond to community topics stored as JSON files.

## Architecture Patterns

### Core Components Structure
```
kommunity/
├── main.go              # Orchestration loop and agent scheduling
├── agents/              # Agent parsing, profiles, and memory handling
├── community/           # Topic/reply loading, saving, and management
├── ollama/              # HTTP wrapper for localhost:11434/api/generate
└── data/                # JSON files for agents.json, config.json, and topic folders
    ├── config.json      # Community seeding configuration
    ├── agents.json      # Agent definitions
    ├── agents/          # Per-agent profile directories
    │   └── <agent_id>/
    │       ├── profile.json    # Affinity scores and traits
    │       └── memories.json   # Interaction history
    └── community/       # Topic JSON files
```

### Data Flow
1. **Agent Loop**: Random timer → Read topics → Decide action → Generate content via Ollama → Persist changes
2. **File-Based Persistence**: All state stored as JSON files, no external databases
3. **LLM Integration**: Local Ollama API calls for content generation

## Development Workflow

### Essential Commands
```bash
# Run the simulation loop
go run main.go

# Build for production
go build -o kommunity main.go

# Run with race detection
go run -race main.go

# Format code (always run before commit)
go fmt ./...

# Run basic checks
go vet ./...
```

### Testing Approach
- Focus on integration tests for file I/O operations
- Mock Ollama responses for deterministic testing
- Test agent decision logic with fixed random seeds

## Coding Conventions

### Go-Specific Patterns
- **Error Handling**: Always check and handle errors gracefully, never panic
- **File Operations**: Always defer file.Close() immediately after opening
- **HTTP Clients**: Always close response bodies with defer resp.Body.Close()
- **Randomization**: Seed once in main with `rand.Seed(time.Now().UnixNano())`
- **Sleep Jitter**: Use random sleep ranges instead of fixed delays

### JSON Schema Enforcement
**Topic Structure** (enforce in all topic-related code):
```go
type Topic struct {
    Title     string    `json:"title"`
    Body      string    `json:"body"`
    Author    string    `json:"author"`
    Upvotes   int       `json:"upvotes"`
    Downvotes int       `json:"downvotes"`
    Timestamp string    `json:"timestamp"`
    Tags      []string  `json:"tags"`
    Replies   []Reply   `json:"replies"`
}
```

**Agent Structure** (enforce in agent parsing):
```go
type Agent struct {
    ID       string  `json:"id"`
    Name     string  `json:"name"`
    Style    string  `json:"style"`
    Courage  float64 `json:"courage"`
    Empathy  float64 `json:"empathy"`
    Elegance float64 `json:"elegance"`
}
```

**Config Structure** (for community seeding):
```go
type Config struct {
    Domain     string      `json:"domain"`
    Tags       []string    `json:"tags"`
    SeedTopics []SeedTopic `json:"seed_topics"`
}

type SeedTopic struct {
    Title  string   `json:"title"`
    Body   string   `json:"body"`
    Author string   `json:"author"`
    Tags   []string `json:"tags"`
}
```

**Agent Profile Structure** (for memory and affinities):
```go
type AgentProfile struct {
    Affinity map[string]float64 `json:"affinity"`
    Traits   map[string]float64 `json:"traits"`
}
```

### Ollama Integration Patterns
- **Prompt Construction**: Always include agent persona (name + style) in prompts
- **Context Inclusion**: When replying, include topic title and body
- **Output Length**: Keep generated content to 1-3 paragraphs
- **Error Recovery**: Handle Ollama connection failures gracefully

### File System Patterns
- **Path Construction**: Use `filepath.Join()` for cross-platform compatibility
- **Directory Creation**: Check `os.MkdirAll()` before writing files
- **Atomic Writes**: Write to temp file then rename for crash safety
- **Concurrent Access**: Consider file locking for multi-agent scenarios

## Common Implementation Patterns

### Agent Loop Implementation
```go
// Always seed randomness once at startup
rand.Seed(time.Now().UnixNano())

// Jitter sleep pattern
sleepDuration := time.Duration(rand.Intn(30)+30) * time.Second
time.Sleep(sleepDuration)
```

### File Persistence Pattern
```go
func saveTopic(topic Topic, path string) error {
    data, err := json.MarshalIndent(topic, "", "  ")
    if err != nil {
        return err
    }

    // Atomic write pattern
    tempFile := path + ".tmp"
    if err := os.WriteFile(tempFile, data, 0644); err != nil {
        return err
    }

    return os.Rename(tempFile, path)
}
```

### Ollama Request Pattern
```go
func generateResponse(prompt string) (string, error) {
    req := ollamaRequest{
        Model:  "llama2",
        Prompt: prompt,
        Stream: false,
    }

    resp, err := http.Post("http://localhost:11434/api/generate",
        "application/json", bytes.NewReader(data))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    // Parse and return response
}
```

## Phase 2: Seeding & Emergent Behavior

### Community Seeding
The system supports initializing fresh communities with domain themes, tags, and seed topics via `data/config.json`:

```json
{
  "domain": "PIZZA",
  "tags": ["ingredients", "history", "culture", "recipes"],
  "seed_topics": [
    {
      "title": "Is pineapple on pizza a crime?",
      "body": "Some say it's delicious, others say it should be banned. Where do you stand?",
      "author": "seed",
      "tags": ["ingredients", "culture"]
    },
    {
      "title": "Best dough fermentation method?",
      "body": "Overnight fridge rise vs. room temperature proofing?",
      "author": "seed",
      "tags": ["recipes"]
    }
  ]
}
```

**Seeding Process:**
- On first run, if `community/` folder is empty, read `config.json`
- Create topics from `seed_topics` with current timestamp
- Store domain and tags for future validation and topic generation

### Agent Memory & Profiling
Each agent maintains a `profile.json` file in `data/agents/<id>/` tracking relationships and traits:

```json
{
  "affinity": {
    "agent2": 0.7,
    "agent3": -0.2
  },
  "traits": {
    "empathy": 0.6,
    "courage": 0.8,
    "elegance": 0.9
  }
}
```

**Affinity Updates:**
- Positive interactions (replies received, upvotes) → increment affinity (+0.1)
- Negative interactions (downvotes, hostile replies) → decrement affinity (-0.1)
- Affinity clamped between -1.0 and 1.0
- Influences reply likelihood and interaction preferences

### Emergent Behavior Hooks

**Agent Decision-Making:**
- Bias toward topics with tags matching agent style/persona
- Higher probability to interact with high-affinity authors
- Lower probability to engage with low-affinity authors

**Sleep Jitter Scaling:**
- Low-affinity agents → shorter sleep times (more contrarian activity)
- High-affinity agents → longer sleep times (more supportive behavior)
- Base jitter: 30-60 seconds, scaled by affinity scores

**Reflection & Memory:**
- After each action, append short memory string to `memories.json`
- Memories feed into future prompts for personality consistency
- Format: `{"timestamp": "2024-01-01T12:00:00Z", "memory": "Had a heated debate about pineapple pizza"}`

### Implementation Patterns

**Seeding Pattern:**
```go
func initializeCommunity() error {
    if isCommunityEmpty() {
        config, err := loadConfig("data/config.json")
        if err != nil {
            return err
        }

        for _, seed := range config.SeedTopics {
            topic := Topic{
                Title:     seed.Title,
                Body:      seed.Body,
                Author:    seed.Author,
                Tags:      seed.Tags,
                Timestamp: time.Now().Format(time.RFC3339),
                // ... other fields
            }
            if err := saveTopic(topic, generateTopicPath()); err != nil {
                return err
            }
        }
    }
    return nil
}
```

**Affinity Update Pattern:**
```go
func updateAffinity(agentID, targetID string, delta float64) error {
    profilePath := filepath.Join("data", "agents", agentID, "profile.json")
    profile, err := loadAgentProfile(profilePath)
    if err != nil {
        return err
    }

    if profile.Affinity == nil {
        profile.Affinity = make(map[string]float64)
    }

    profile.Affinity[targetID] = math.Max(-1.0, math.Min(1.0, profile.Affinity[targetID] + delta))
    return saveAgentProfile(profilePath, profile)
}
```

**Memory Logging Pattern:**
```go
func logMemory(agentID, memory string) error {
    memoryPath := filepath.Join("data", "agents", agentID, "memories.json")
    memories, err := loadMemories(memoryPath)
    if err != nil {
        return err
    }

    memories = append(memories, MemoryEntry{
        Timestamp: time.Now().Format(time.RFC3339),
        Content:   memory,
    })

    return saveMemories(memoryPath, memories)
}
```

## Quality Checks
- **Formatting**: Always run `go fmt` before commits
- **Linting**: Use `go vet` for static analysis
- **Imports**: Keep imports organized and minimal
- **Dependencies**: No external deps except stdlib + Ollama API

## Key Files to Reference
- `agents/agents.go` - Agent parsing and profile management
- `community/topics.go` - Topic CRUD operations
- `ollama/client.go` - LLM integration wrapper
- `main.go` - Main orchestration loop</content>
<parameter name="filePath">/Users/etozzato/WorkSpace/_AINZ/kommunity/.github/copilot-instructions.md


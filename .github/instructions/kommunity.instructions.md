---
applyTo: '**'
---
# Kommunity Engineering Instructions

## Context
We are building a lightweight, file-based community simulator:
- **Agents** are personas configured via `agents.json`.
- **Community** is represented as a folder of JSON topic files.
- **Agent loop**: each agent wakes on a random timer, reads topics, decides whether to reply or create a new one, then persists changes.
- **LLM Backend**: we use Ollama locally to generate agent text.

## Coding Guidelines
1. **Language**: Go (>=1.22).
2. **Filesystem over DB**: all persistence is JSON files in folders, no external DB.
3. **Modularity**:
   - `agents/` → parsing, profile/memory handling.
   - `community/` → topic/reply loading/saving.
   - `ollama/` → wrapper around `http://localhost:11434/api/generate`.
   - `main.go` → orchestration loop.
4. **Error Handling**:
   - Fail gracefully, log instead of panic.
   - Always close files and HTTP responses.
5. **Determinism**:
   - Use `rand.Seed(time.Now().UnixNano())` once in `main`.
   - All random sleeps should be jittered within a range.
6. **Extensibility**:
   - Agents may later store `profile.json` with affinities toward other agents.
   - Community topics may include votes, timestamps, replies.
7. **Prompts to Ollama**:
   - Always include agent persona (name + style).
   - When replying, include topic title and body for context.
   - Keep outputs short (1–3 paragraphs).
8. **JSON Formats**:
   - Topic JSON must include:
     ```json
     {
       "title": "...",
       "body": "...",
       "author": "...",
       "upvotes": 0,
       "downvotes": 0,
       "timestamp": "...",
       "replies": []
     }
     ```
   - Agents JSON must include:
     ```json
     {
       "id": "agent1",
       "name": "Plato",
       "style": "philosopher, reflective, loves analogies",
       "courage": 0.8,
       "empathy": 0.6,
       "elegance": 0.9
     }
     ```

## Development Workflow
- Use `go run main.go` for local loop.
- Commit small, testable chunks.
- Prefer simplicity over optimization; this is a simulation, not prod infra.
- Keep everything self-contained (no external deps except stdlib + Ollama API).

---
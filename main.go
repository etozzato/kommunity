package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"kommunity/agents"
	"kommunity/community"
	"kommunity/ollama"
)

func main() {
	serve := flag.Bool("serve", false, "start the web interface")
	addr := flag.String("addr", ":8080", "address for the web interface")
	flag.Parse()

	if *serve {
		if err := runServer(*addr); err != nil {
			log.Fatalf("failed to start web server: %v", err)
		}
		return
	}

	fmt.Println("🚀 Starting Kommunity Simulator...")

	// Load agents
	agentList, err := agents.LoadAgents("data/agents.json")
	if err != nil {
		fmt.Printf("Error loading agents: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d agents\n", len(agentList))

	// Initialize community if empty
	if err := community.InitializeIfEmpty("data/config.json"); err != nil {
		fmt.Printf("Error initializing community: %v\n", err)
		return
	}

	// Main simulation loop
	fmt.Println("🎭 Simulation starting... (Ctrl+C to stop)")
	for {
		// Select random agent
		agent := agentList[rand.Intn(len(agentList))]

		// Agent performs action
		if err := performAgentAction(agent); err != nil {
			fmt.Printf("Agent %s error: %v\n", agent.Name, err)
		}

		// Sleep with jitter
		// sleepDuration := time.Duration(rand.Intn(30)+30) * time.Second
		sleepDuration := 5 * time.Second
		time.Sleep(sleepDuration)
	}
}

func performAgentAction(agent agents.Agent) error {
	fmt.Printf("🤖 %s (%s) is thinking...\n", agent.Name, agent.Style)

	// Load recent topics
	topics, err := community.LoadRecentTopics("data/community", 5)
	if err != nil {
		return fmt.Errorf("loading topics: %w", err)
	}

	fmt.Printf("   📚 Found %d recent topics\n", len(topics))

	// Decide action (simplified for now)
	action := decideAction(agent, topics)
	fmt.Printf("   🎯 Decided to: %s\n", action)

	switch action {
	case "create_topic":
		return createNewTopic(agent)
	case "reply":
		if len(topics) > 0 {
			// Select a random topic from recent ones to encourage broader participation
			selectedTopic := topics[rand.Intn(len(topics))]
			fmt.Printf("   🎲 Selected topic for reply: '%s' (by %s)\n", selectedTopic.Title[:min(50, len(selectedTopic.Title))]+"...", selectedTopic.Author)
			return replyToTopic(agent, selectedTopic)
		}
	}

	return nil
}

func decideAction(agent agents.Agent, topics []community.Topic) string {
	// Enhanced decision logic - 15% chance to create, 85% to reply if topics exist
	// This encourages more conversation depth
	if len(topics) == 0 || rand.Float64() < 0.15 {
		return "create_topic"
	}
	return "reply"
}

func createNewTopic(agent agents.Agent) error {
	prompt := fmt.Sprintf("You are %s, %s. Create an interesting discussion topic for our community. Keep it to 1-2 sentences.", agent.Name, agent.Style)

	fmt.Printf("   📝 Sending prompt to Ollama: %s\n", prompt[:min(100, len(prompt))]+"...")

	content, err := ollama.GenerateResponse(prompt)
	if err != nil {
		return fmt.Errorf("generating topic: %w", err)
	}

	fmt.Printf("   ✨ Generated topic: %s\n", content[:min(100, len(content))]+"...")

	topic := community.Topic{
		Title:     content,
		Body:      content,
		Author:    agent.ID,
		Timestamp: time.Now().Format(time.RFC3339),
		Tags:      []string{}, // Will be enhanced in Phase 2
		Replies:   []community.Reply{},
	}

	if err := community.SaveTopic(topic, "data/community"); err != nil {
		return fmt.Errorf("saving topic: %w", err)
	}

	fmt.Printf("   💾 Topic saved successfully\n")
	return nil
}

func replyToTopic(agent agents.Agent, topic community.Topic) error {
	// Build conversation context
	context := fmt.Sprintf("Original Topic: %s\n\n%s", topic.Title, topic.Body)

	if len(topic.Replies) > 0 {
		context += "\n\nPrevious Replies:\n"
		for i, reply := range topic.Replies {
			context += fmt.Sprintf("%d. %s: %s\n", i+1, reply.Author, reply.Content)
		}
	}

	prompt := fmt.Sprintf("You are %s, %s. Here is the ongoing discussion:\n\n%s\n\nPlease provide a thoughtful reply that adds value to this conversation. Keep your response to 1-2 sentences.", agent.Name, agent.Style, context)

	fmt.Printf("   💬 Replying to topic with %d existing replies\n", len(topic.Replies))
	fmt.Printf("   📝 Sending prompt to Ollama: %s\n", prompt[:min(150, len(prompt))]+"...")

	content, err := ollama.GenerateResponse(prompt)
	if err != nil {
		return fmt.Errorf("generating reply: %w", err)
	}

	fmt.Printf("   ✨ Generated reply: %s\n", content[:min(100, len(content))]+"...")

	reply := community.Reply{
		Author:    agent.ID,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err := community.AddReplyToTopic(topic.Title, reply, "data/community"); err != nil {
		return fmt.Errorf("adding reply: %w", err)
	}

	fmt.Printf("   💾 Reply saved successfully\n")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

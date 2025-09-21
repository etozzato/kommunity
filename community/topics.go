package community

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Topic represents a discussion topic
type Topic struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	Author    string   `json:"author"`
	Upvotes   int      `json:"upvotes"`
	Downvotes int      `json:"downvotes"`
	Timestamp string   `json:"timestamp"`
	Tags      []string `json:"tags"`
	Replies   []Reply  `json:"replies"`
	Filename  string   `json:"-"`
}

// Reply represents a reply to a topic
type Reply struct {
	Author    string `json:"author"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// Config represents community configuration for seeding
type Config struct {
	Domain     string      `json:"domain"`
	Tags       []string    `json:"tags"`
	SeedTopics []SeedTopic `json:"seed_topics"`
}

// SeedTopic represents a seed topic for initialization
type SeedTopic struct {
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Author string   `json:"author"`
	Tags   []string `json:"tags"`
}

// LoadRecentTopics loads the most recent topics from the community directory
func LoadRecentTopics(dir string, limit int) ([]Topic, error) {
	topics, err := LoadTopics(dir)
	if err != nil {
		return nil, err
	}
	if limit > 0 && len(topics) > limit {
		topics = topics[:limit]
	}
	return topics, nil
}

func LoadTopics(dir string) ([]Topic, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving community directory: %w", err)
	}

	var topics []Topic
	if err := filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}
		topic, err := loadTopic(path)
		if err != nil {
			return nil // skip corrupted files silently
		}
		if rel, relErr := filepath.Rel(absDir, path); relErr == nil {
			topic.Filename = rel
		}
		topics = append(topics, topic)
		return nil
	}); err != nil {
		if os.IsNotExist(err) {
			return []Topic{}, nil
		}
		return nil, fmt.Errorf("walking community directory: %w", err)
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Timestamp > topics[j].Timestamp
	})

	return topics, nil
}

// SaveTopic saves a topic to the community directory
func SaveTopic(topic Topic, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving community directory: %w", err)
	}

	var path string
	if topic.Filename != "" {
		path = topic.Filename
		if !filepath.IsAbs(path) {
			path = filepath.Join(absDir, path)
		}
	} else {
		// Create directory if it doesn't exist
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("creating community directory: %w", err)
		}

		// Generate filename from title (simplified)
		filename := strings.ReplaceAll(strings.ToLower(topic.Title), " ", "_")
		filename = strings.ReplaceAll(filename, "'", "")
		filename = fmt.Sprintf("%s.json", filename[:min(50, len(filename))])

		path = filepath.Join(absDir, filename)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating topic directory: %w", err)
	}

	data, err := json.MarshalIndent(topic, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling topic: %w", err)
	}

	// Atomic write
	tempFile := path + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := os.Rename(tempFile, path); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	if rel, relErr := filepath.Rel(absDir, path); relErr == nil {
		topic.Filename = rel
	} else {
		topic.Filename = path
	}

	return nil
}

// AddReplyToTopic adds a reply to an existing topic
func AddReplyToTopic(topicTitle string, reply Reply, dir string) error {
	topics, err := LoadTopics(dir)
	if err != nil {
		return err
	}

	for _, topic := range topics {
		if topic.Title == topicTitle {
			topic.Replies = append(topic.Replies, reply)
			return SaveTopic(topic, dir)
		}
	}

	return fmt.Errorf("topic not found: %s", topicTitle)
}

func LoadTopicByRelativePath(dir, relPath string) (Topic, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return Topic{}, fmt.Errorf("resolving community directory: %w", err)
	}

	clean := filepath.Clean(relPath)
	if strings.HasPrefix(clean, "..") {
		return Topic{}, fmt.Errorf("invalid topic path: %s", relPath)
	}
	path := filepath.Join(absDir, clean)
	topic, err := loadTopic(path)
	if err != nil {
		return Topic{}, err
	}
	if rel, relErr := filepath.Rel(absDir, path); relErr == nil {
		topic.Filename = rel
	}
	return topic, nil
}

// InitializeIfEmpty initializes the community with seed topics if empty
func InitializeIfEmpty(configPath string) error {
	// Check if community directory is empty
	files, err := os.ReadDir("data/community")
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, create it
			if err := os.MkdirAll("data/community", 0755); err != nil {
				return fmt.Errorf("creating community directory: %w", err)
			}
		} else {
			return fmt.Errorf("reading community directory: %w", err)
		}
	}

	if len(files) > 0 {
		// Community already has topics
		return nil
	}

	// Load config and seed topics
	config, err := loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Printf("🌱 Seeding community with %d topics for domain: %s\n", len(config.SeedTopics), config.Domain)

	for _, seed := range config.SeedTopics {
		topic := Topic{
			Title:     seed.Title,
			Body:      seed.Body,
			Author:    seed.Author,
			Tags:      seed.Tags,
			Timestamp: time.Now().Format(time.RFC3339),
			Upvotes:   0,
			Downvotes: 0,
			Replies:   []Reply{},
		}

		if err := SaveTopic(topic, "data/community"); err != nil {
			return fmt.Errorf("saving seed topic: %w", err)
		}
	}

	return nil
}

func loadTopic(path string) (Topic, error) {
	file, err := os.Open(path)
	if err != nil {
		return Topic{}, fmt.Errorf("opening topic file: %w", err)
	}
	defer file.Close()

	var topic Topic
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&topic); err != nil {
		return Topic{}, fmt.Errorf("decoding topic JSON: %w", err)
	}
	topic.Filename = path

	return topic, nil
}

func loadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("opening config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("decoding config JSON: %w", err)
	}

	return config, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

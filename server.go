package main

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"kommunity/community"
)

type topicSummary struct {
	Title      string
	Author     string
	Timestamp  string
	When       string
	Snippet    string
	Tags       []string
	ReplyCount int
	Path       string
}

type topicDetail struct {
	Title     string
	Body      string
	Author    string
	Timestamp string
	When      string
	Tags      []string
	Replies   []community.Reply
}

func runServer(addr string) error {
	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"formatTime": formatTime,
	})
	router.LoadHTMLGlob("web/templates/*.tmpl")

	router.GET("/", func(c *gin.Context) {
		topics, err := community.LoadTopics("data/community")
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to load topics: %v", err)
			return
		}

		summaries := make([]topicSummary, 0, len(topics))
		for _, t := range topics {
			summaries = append(summaries, topicSummary{
				Title:      t.Title,
				Author:     t.Author,
				Timestamp:  t.Timestamp,
				When:       formatTime(t.Timestamp),
				Snippet:    buildSnippet(t.Body),
				Tags:       t.Tags,
				ReplyCount: len(t.Replies),
				Path:       toURLPath(t.Filename),
			})
		}

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Topics": summaries,
			"Count":  len(summaries),
		})
	})

	router.GET("/topic/*topicPath", func(c *gin.Context) {
		rel := strings.TrimPrefix(c.Param("topicPath"), "/")
		if rel == "" {
			c.Redirect(http.StatusFound, "/")
			return
		}

		topic, err := community.LoadTopicByRelativePath("data/community", rel)
		if err != nil {
			c.String(http.StatusNotFound, "topic not found: %v", err)
			return
		}

		detail := topicDetail{
			Title:     topic.Title,
			Body:      topic.Body,
			Author:    topic.Author,
			Timestamp: topic.Timestamp,
			When:      formatTime(topic.Timestamp),
			Tags:      topic.Tags,
			Replies:   topic.Replies,
		}

		c.HTML(http.StatusOK, "topic.tmpl", gin.H{
			"Topic":    detail,
			"FilePath": filepath.ToSlash(topic.Filename),
			"LinkPath": toURLPath(topic.Filename),
		})
	})

	return router.Run(addr)
}

func formatTime(ts string) string {
	if ts == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, ts); err == nil {
		return parsed.Format("Jan 2, 2006 15:04 MST")
	}
	if parsed, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return parsed.Format("Jan 2, 2006 15:04 MST")
	}
	return ts
}

func buildSnippet(body string) string {
	trimmed := strings.TrimSpace(body)
	runes := []rune(trimmed)
	if len(runes) <= 160 {
		return trimmed
	}
	return string(runes[:157]) + "..."
}

func toURLPath(rel string) string {
	if rel == "" {
		return ""
	}
	return "/topic/" + strings.TrimPrefix(filepath.ToSlash(rel), "/")
}

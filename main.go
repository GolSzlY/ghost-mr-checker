package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"ghost-mr-checker/checker"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v3"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	data, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	if cfg.GitLab.Token == "" || cfg.GitLab.ProjectID == "" {
		log.Fatal("GitLab token and project ID are required in config")
	}

	sinceStr := cfg.Check.Since
	if sinceStr == "" {
		sinceStr = "2025-11-09"
	}

	since, err := time.Parse("2006-01-02", sinceStr)
	if err != nil {
		log.Fatalf("Invalid date format: %v", err)
	}

	baseURL := cfg.GitLab.URL
	if baseURL == "" {
		baseURL = "https://gitlab.myworkplaze.com/api/v4"
	}

	git, err := gitlab.NewClient(cfg.GitLab.Token, gitlab.WithBaseURL(baseURL))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	c := checker.NewChecker(git.Commits, git.MergeRequests, cfg.GitLab.ProjectID)

	results, err := c.Check(since)
	if err != nil {
		log.Fatalf("Check failed: %v", err)
	}

	for _, res := range results {
		fmt.Printf("[%s] %s (Source: %s)\n", res.Status, res.ReleaseMR.Title, res.ReleaseMR.SourceBranch)
	}
}

type Config struct {
	GitLab struct {
		Token     string `yaml:"token"`
		ProjectID string `yaml:"project_id"`
		URL       string `yaml:"url"`
	} `yaml:"gitlab"`
	Check struct {
		Since string `yaml:"since"`
	} `yaml:"check"`
}

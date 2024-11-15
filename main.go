package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed instructions/*
var instructions embed.FS

const ClaudeApiUrl = "https://api.anthropic.com/v1/messages"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages,omitempty"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type Response struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Error struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func loadServiceTemplates(dockerComposeContent string) (string, error) {
	entries, err := instructions.ReadDir("instructions")
	if err != nil {
		return "", fmt.Errorf("error reading instructions directory: %v", err)
	}

	var services []string
	for _, entry := range entries {
		serviceName := strings.TrimSuffix(entry.Name(), ".md")
		if strings.Contains(strings.ToLower(dockerComposeContent), strings.ToLower(serviceName)) {
			content, err := instructions.ReadFile(path.Join("instructions", entry.Name()))
			if err != nil {
				return "", fmt.Errorf("error reading template for %s: %v", serviceName, err)
			}
			services = append(services, fmt.Sprintf(`<service>
<name>%s</name>
<template>%s</template>
</service>`, serviceName, string(content)))
		}
	}

	if len(services) > 0 {
		return fmt.Sprintf("<services>%s</services>", strings.Join(services, "\n")), nil
	}
	return "", nil
}

func callClaude(apiKey string, dockerCompose string, schema string) (string, error) {
	serviceDocs, err := loadServiceTemplates(dockerCompose)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`<input>
<docker-compose>%s</docker-compose>
<schema>%s</schema>
%s
<instructions>
1. Convert the docker-compose.yaml to zeabur-template.yaml based on the provided schema.
2. Use provided service templates directly when available.
3. Place config content directly in YAML instead of using volume mounts.
  Exception: configs that auto-generate at startup and reset on restart.
</instructions>
</input>
<output-format>
Provide only zeabur-template.yaml content without explanations or code blocks.
</output-format>
`, dockerCompose, schema, serviceDocs)

	requestBody := Request{
		Model:     "claude-3-sonnet-20240229",
		Messages:  []Message{{Role: "user", Content: prompt}},
		MaxTokens: 4096,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	req, err := http.NewRequest("POST", ClaudeApiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error unmarshaling response: %v", err)
	}

	if response.Error.Message != "" {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Content) > 0 {
		return response.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content in response")
}

func main() {
	apiKey := os.Getenv("CLAUDE_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set CLAUDE_API_KEY environment variable")
		return
	}

	dockerCompose, err := os.ReadFile("docker-compose.yaml")
	if err != nil {
		fmt.Printf("Error reading docker-compose.yaml: %v\n", err)
		return
	}

	schema, err := os.ReadFile("schema.json")
	if err != nil {
		fmt.Printf("Error reading schema.json: %v\n", err)
		return
	}

	zeaburTemplate, err := callClaude(apiKey, string(dockerCompose), string(schema))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	err = os.WriteFile("zeabur-template.yaml", []byte(zeaburTemplate), 0644)
	if err != nil {
		fmt.Printf("Error writing zeabur-template.yaml: %v\n", err)
		return
	}

	fmt.Println("Successfully converted to zeabur-template.yaml")
}

package distill

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/eduardmaghakyan/gitlore/internal/config"
)

const (
	timeout    = 30 * time.Second
	maxDiff    = 4000
	apiURL     = "https://api.anthropic.com/v1/messages"
	apiVersion = "2023-06-01"
	fallback   = "[gitlore] Conversation captured but distillation unavailable."
)

// Distill produces a 3-sentence summary from the conversation and diff.
func Distill(conversation, diff string, cfg *config.Config) string {
	diff = truncateDiff(diff)
	prompt := fmt.Sprintf(promptTemplate, conversation, diff)

	if cfg.Distill.UseCLI {
		if result, err := viaCLI(prompt, cfg.Distill.Model); err == nil {
			return result
		}
	}

	if cfg.Distill.APIKey != "" {
		if result, err := viaAPI(prompt, cfg); err == nil {
			return result
		}
	}

	return fallback
}

func viaCLI(prompt, model string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := []string{"-p", "--output-format", "text"}
	if model != "" {
		args = append(args, "--model", model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = strings.NewReader(prompt)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(out))
	if result == "" {
		return "", fmt.Errorf("empty response")
	}
	return result, nil
}

func viaAPI(prompt string, cfg *config.Config) (string, error) {
	body := map[string]interface{}{
		"model":      cfg.Distill.Model,
		"max_tokens": 300,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.Distill.APIKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API returned %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return "", err
	}

	if len(apiResp.Content) == 0 || apiResp.Content[0].Text == "" {
		return "", fmt.Errorf("empty API response")
	}

	return strings.TrimSpace(apiResp.Content[0].Text), nil
}

func truncateDiff(diff string) string {
	if len(diff) <= maxDiff {
		return diff
	}
	// Keep the first part (most relevant changes) and a stat summary
	return diff[:maxDiff] + "\n... (diff truncated)"
}

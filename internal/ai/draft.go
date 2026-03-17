package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

type apiRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type apiResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// GenerateDraft calls the Anthropic Messages API and returns the generated ADR draft sections.
func GenerateDraft(prompt, model, apiKey string) (string, error) {
	return generateDraft(prompt, model, apiKey, http.DefaultClient)
}

func generateDraft(prompt, model, apiKey string, client *http.Client) (string, error) {
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY is not set.\nPlease set the environment variable and try again.\n\n  export ANTHROPIC_API_KEY=your-api-key")
	}

	reqBody := apiRequest{
		Model:     model,
		MaxTokens: 2048,
		Messages:  []message{{Role: "user", Content: prompt}},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("Anthropic API error: %s", apiResp.Error.Message)
	}

	for _, c := range apiResp.Content {
		if c.Type == "text" {
			return c.Text, nil
		}
	}
	return "", fmt.Errorf("no text content in API response")
}

// DryRun prints the prompt to w and asks the user to confirm via r.
// Returns true if the user confirms (enters "y").
func DryRun(prompt string, r io.Reader, w io.Writer) bool {
	fmt.Fprintln(w, "--- Prompt to be sent to Anthropic API ---")
	fmt.Fprint(w, prompt)
	fmt.Fprintln(w, "------------------------------------------")
	fmt.Fprint(w, "Proceed? (y/n): ")

	scanner := bufio.NewScanner(r)
	if !scanner.Scan() {
		return false
	}
	return strings.TrimSpace(strings.ToLower(scanner.Text())) == "y"
}

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type bedrockRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []message `json:"messages"`
}

// GenerateDraftBedrock calls AWS Bedrock InvokeModel and returns the generated draft.
// AWS credentials and region are resolved from the standard environment
// (AWS_PROFILE, AWS_REGION, AWS_ACCESS_KEY_ID, etc.).
func GenerateDraftBedrock(prompt, model string) (string, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return "", fmt.Errorf("loading AWS config: %w", err)
	}
	if cfg.Region == "" {
		return "", fmt.Errorf("AWS region is not set.\nSet AWS_REGION or configure a default region in your AWS profile.")
	}

	client := bedrockruntime.NewFromConfig(cfg)

	reqBody := bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        2048,
		Messages:         []message{{Role: "user", Content: prompt}},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("encoding request: %w", err)
	}

	resp, err := client.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(model),
		Body:        body,
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return "", fmt.Errorf("calling Bedrock: %w", err)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(bytes.NewReader(resp.Body)).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("decoding Bedrock response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("Bedrock error: %s", apiResp.Error.Message)
	}

	for _, c := range apiResp.Content {
		if c.Type == "text" {
			return c.Text, nil
		}
	}
	return "", fmt.Errorf("no text content in Bedrock response")
}

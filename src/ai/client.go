package ai

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// ClientConfig はAIクライアントの設定
type ClientConfig struct {
	APIKey string
}

// NewGenaiClient は新しいGenai clientを作成する共通ファクトリー
func NewGenaiClient(ctx context.Context) (*genai.Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return client, nil
}

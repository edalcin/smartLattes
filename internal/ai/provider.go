package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrInvalidKey          = errors.New("chave de API inválida")
	ErrProviderUnavailable = errors.New("provedor de IA indisponível")
	ErrTimeout             = errors.New("tempo limite excedido")
	ErrRateLimited         = errors.New("limite de requisições atingido")
)

type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type GenerateRequest struct {
	APIKey       string
	Model        string
	SystemPrompt string
	UserData     string
	MaxTokens    int
}

type AIProvider interface {
	ListModels(ctx context.Context, apiKey string) ([]Model, error)
	Generate(ctx context.Context, req GenerateRequest) (string, error)
}

type OpenAIProvider struct{}
type AnthropicProvider struct{}
type GeminiProvider struct{}

// extractAPIError attempts to parse the provider's JSON error response body
// and return a human-readable message. Falls back to raw body if parsing fails.
func extractAPIError(body []byte, provider string) string {
	if len(body) == 0 {
		return provider + ": sem detalhes"
	}
	// Try OpenAI/Anthropic format: {"error": {"message": "..."}}
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
		return errResp.Error.Message
	}
	// Truncate raw body to avoid flooding the UI
	s := string(body)
	if len(s) > 300 {
		s = s[:300] + "..."
	}
	return s
}

func NewProvider(name string) (AIProvider, error) {
	switch name {
	case "openai":
		return &OpenAIProvider{}, nil
	case "anthropic":
		return &AnthropicProvider{}, nil
	case "gemini":
		return &GeminiProvider{}, nil
	default:
		return nil, fmt.Errorf("provedor desconhecido: %s", name)
	}
}

package ai

import (
	"context"
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

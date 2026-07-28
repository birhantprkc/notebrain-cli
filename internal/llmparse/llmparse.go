package llmparse

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// ErrNoAPIKey is returned when no valid API key is found for LLM parsing.
var ErrNoAPIKey = errors.New("no valid API key found")

// API Endpoints
const (
	EndpointDeepSeek   = "https://api.deepseek.com"
	EndpointOpenRouter = "https://openrouter.ai/api/v1"
	EndpointOpenAI     = "https://api.openai.com/v1"
	EndpointGemini     = "https://generativelanguage.googleapis.com/v1beta/openai/"
	EndpointOllama     = "http://localhost:11434/v1"
)

const backendOllama = "ollama"

type BackendConfig struct {
	Name    string
	BaseURL string
	EnvKey  string
	Headers map[string]string
}

var supportedBackends = []BackendConfig{
	{Name: "deepseek", BaseURL: EndpointDeepSeek, EnvKey: "DEEPSEEK_API_KEY"},
	{
		Name:    "openrouter",
		BaseURL: EndpointOpenRouter,
		EnvKey:  "OPENROUTER_API_KEY",
		Headers: map[string]string{
			"HTTP-Referer": "https://github.com/nmdra/notebrain-cli",
			"X-Title":      "NoteBrain CLI",
		},
	},
	{Name: "openai", BaseURL: EndpointOpenAI, EnvKey: "OPENAI_API_KEY"},
	{Name: "gemini", BaseURL: EndpointGemini, EnvKey: "GEMINI_API_KEY"},
	{Name: backendOllama, BaseURL: EndpointOllama, EnvKey: "OLLAMA_API_KEY"},
}

// Converter converts raw PDF page text into structured Markdown via an LLM.
type Converter interface {
	// Convert takes raw text pages and returns a single Markdown string.
	Convert(ctx context.Context, pages []string) (string, error)
	// Name returns the backend name for logging (e.g. "gemini", "openrouter", "deepseek").
	Name() string
}

// New creates a Converter for the given model name.
func New(model string, contextWindow int) (Converter, error) {
	if contextWindow < 12288 { // 8192 reserve + 4096 min chunk size
		return nil, fmt.Errorf("context window %d is too small (needs at least 12288 tokens)", contextWindow)
	}

	var selected *BackendConfig
	var apiKey string
	var baseURL string

	// Auto-detect based on API key presence
	for i := range supportedBackends {
		b := &supportedBackends[i]
		val := os.Getenv(b.EnvKey)
		// Properly trim outside whitespace first, then cutset characters
		val = strings.Trim(strings.TrimSpace(val), "\"'\t\n")
		val = strings.ReplaceAll(val, " ", "")

		// Special case for Ollama keyless API
		if b.Name == backendOllama && val == "" {
			if host := os.Getenv("OLLAMA_HOST"); host != "" {
				val = "dummy"
				baseURL = host // Override endpoint with custom host
			}
		}

		if val != "" {
			selected = b
			apiKey = val
			if baseURL == "" {
				baseURL = b.BaseURL
			}
			break
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("%w for auto-detection (checked DEEPSEEK_API_KEY, OPENROUTER_API_KEY, OPENAI_API_KEY, GEMINI_API_KEY, OLLAMA_API_KEY/OLLAMA_HOST) for model %s", ErrNoAPIKey, model)
	}

	actualModel := strings.TrimPrefix(model, "openrouter/")

	slog.Info("using LLM backend", "backend", selected.Name, "model", actualModel)
	return newOpenAICompatConverter(baseURL, apiKey, actualModel, selected.Name, contextWindow, selected.Headers), nil
}

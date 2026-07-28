package llmparse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestNewAutoDetection(t *testing.T) {
	os.Clearenv()
	defer os.Clearenv()

	// 1. No keys set -> Error
	_, err := New("some-model", 128000)
	if err == nil {
		t.Errorf("expected error when no API keys are set, got nil")
	}

	// 2. OpenRouter set -> openrouter
	os.Setenv("OPENROUTER_API_KEY", "test-or")
	conv, err := New("some-model", 128000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Name() != "openrouter" {
		t.Errorf("expected backend openrouter, got %s", conv.Name())
	}

	// 3. DeepSeek set -> deepseek (takes precedence over openrouter)
	os.Setenv("DEEPSEEK_API_KEY", "  \"sk-deep seek\"  ")
	conv, err = New("some-model", 128000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conv.Name() != "deepseek" {
		t.Errorf("expected backend deepseek, got %s", conv.Name())
	}

	// Verify key sanitization (space outside and inside quotes)
	cStruct, ok := conv.(*openAICompatConverter)
	if !ok {
		t.Fatalf("expected *openAICompatConverter")
	}
	if cStruct.apiKey != "sk-deepseek" {
		t.Errorf("expected sanitized apiKey 'sk-deepseek', got %q", cStruct.apiKey)
	}

	// 4. Ollama keyless via OLLAMA_HOST when no keys are set
	os.Clearenv()
	os.Setenv("OLLAMA_HOST", "http://localhost:11434")
	conv, err = New("some-model", 128000)
	if err != nil {
		t.Fatalf("unexpected error with OLLAMA_HOST: %v", err)
	}
	if conv.Name() != "ollama" {
		t.Errorf("expected backend ollama, got %s", conv.Name())
	}
}

func TestContextWindowValidation(t *testing.T) {
	os.Setenv("DEEPSEEK_API_KEY", "test")
	defer os.Clearenv()

	_, err := New("some-model", 8000)
	if err == nil {
		t.Errorf("expected error for context window < 12288, got nil")
	}
}

func TestOpenAICompatConverter(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}

		if r.Header.Get("Authorization") != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Model != "test-model" {
			t.Errorf("expected model test-model, got %s", req.Model)
		}

		resp := chatResponse{}
		resp.Choices = append(resp.Choices, struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			Message: struct {
				Content string `json:"content"`
			}{
				Content: "structured markdown",
			},
		})

		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Test trailing slash removal on base URL
	baseURLWithTrailingSlash := mockServer.URL + "/"
	conv := newOpenAICompatConverter(baseURLWithTrailingSlash, "test-key", "test-model", "test-backend", 128000, nil)

	result, err := conv.Convert(context.Background(), []string{"raw text"})
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	if result != "structured markdown" {
		t.Errorf("expected 'structured markdown', got %q", result)
	}
}

func TestOpenAICompatConverter_RetryOn429(t *testing.T) {
	attempts := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": {"message": "rate limit exceeded"}}`))
			return
		}

		resp := chatResponse{}
		resp.Choices = append(resp.Choices, struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			Message: struct {
				Content string `json:"content"`
			}{
				Content: "recovered markdown",
			},
		})
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	conv := newOpenAICompatConverter(mockServer.URL, "test-key", "test-model", "test-backend", 128000, nil)

	result, err := conv.Convert(context.Background(), []string{"raw text"})
	if err != nil {
		t.Fatalf("Convert failed on retry: %v", err)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}

	if result != "recovered markdown" {
		t.Errorf("expected 'recovered markdown', got %q", result)
	}
}

func TestSplitByTokenBudget(t *testing.T) {
	pages := []string{
		strings.Repeat("a", 20), // 20 chars
		strings.Repeat("b", 15), // 15 chars (total 35) -> should fit in chunk 1
		strings.Repeat("c", 10), // 10 chars (total 45) -> exceeds 40, goes to chunk 2
		strings.Repeat("d", 45), // 45 chars -> exceeds 40, single huge page, goes to chunk 3
	}

	chunks := splitByTokenBudget(pages, 10)

	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	if chunks[0] != strings.Repeat("a", 20)+"\n\n---\n\n"+strings.Repeat("b", 15) {
		t.Errorf("unexpected chunk 0: %q", chunks[0])
	}

	if chunks[1] != strings.Repeat("c", 10) {
		t.Errorf("unexpected chunk 1: %q", chunks[1])
	}

	if chunks[2] != strings.Repeat("d", 45) {
		t.Errorf("unexpected chunk 2: %q", chunks[2])
	}
}

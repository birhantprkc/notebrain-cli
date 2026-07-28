package llmparse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	reservedTokens = 8192
	maxRetries     = 5
	initialBackoff = 2 * time.Second
)

type openAICompatConverter struct {
	baseURL       string
	apiKey        string
	model         string
	name          string
	contextWindow int
	headers       map[string]string
}

func newOpenAICompatConverter(baseURL, apiKey, model, name string, contextWindow int, headers map[string]string) *openAICompatConverter {
	return &openAICompatConverter{
		baseURL:       baseURL,
		apiKey:        apiKey,
		model:         model,
		name:          name,
		contextWindow: contextWindow,
		headers:       headers,
	}
}

func (o *openAICompatConverter) Name() string {
	return o.name
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(strings.TrimSpace(header)); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(header); err == nil {
		return time.Until(t)
	}
	return 0
}

func (o *openAICompatConverter) Convert(ctx context.Context, pages []string) (string, error) {
	budget := o.contextWindow - reservedTokens
	if budget < 4096 {
		slog.Warn("calculated chunk budget is extremely small", "budget", budget)
		budget = 4096
	}

	chunks := splitByTokenBudget(pages, budget)

	var results []string
	reqURL := strings.TrimSuffix(o.baseURL, "/") + "/chat/completions"

	for _, chunk := range chunks {
		reqBody := chatRequest{
			Model: o.model,
			Messages: []chatMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: chunk},
			},
			Temperature: 0.1,
		}

		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("%s marshal request: %w", o.name, err)
		}

		var respBody []byte
		var chatResp chatResponse
		backoff := initialBackoff
		success := false

		for attempt := 0; attempt <= maxRetries; attempt++ {
			req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(bodyBytes))
			if err != nil {
				return "", fmt.Errorf("%s create request: %w", o.name, err)
			}

			req.Header.Set("Authorization", "Bearer "+o.apiKey)
			req.Header.Set("Content-Type", "application/json")
			for k, v := range o.headers {
				req.Header.Set(k, v)
			}

			client := &http.Client{Timeout: 90 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				if attempt < maxRetries {
					slog.Warn("request failed, retrying", "backend", o.name, "attempt", attempt+1, "err", err, "backoff", backoff)
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(backoff):
						backoff *= 2
						continue
					}
				}
				return "", fmt.Errorf("%s request failed: %w", o.name, err)
			}

			respBody, err = io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				if attempt < maxRetries {
					slog.Warn("reading response body failed, retrying", "backend", o.name, "attempt", attempt+1, "err", err, "backoff", backoff)
					select {
					case <-ctx.Done():
						return "", ctx.Err()
					case <-time.After(backoff):
						backoff *= 2
						continue
					}
				}
				return "", fmt.Errorf("%s read response: %w", o.name, err)
			}

			chatResp = chatResponse{}
			_ = json.Unmarshal(respBody, &chatResp)

			if resp.StatusCode == http.StatusOK {
				success = true
				break
			}

			if (resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500) && attempt < maxRetries {
				sleepDuration := backoff
				if ra := parseRetryAfter(resp.Header.Get("Retry-After")); ra > 0 {
					sleepDuration = ra
				}
				errMsg := "unknown error"
				if chatResp.Error != nil && chatResp.Error.Message != "" {
					errMsg = chatResp.Error.Message
				} else if len(respBody) > 0 {
					errMsg = strings.TrimSpace(string(respBody))
				}
				slog.Warn("rate limited or server error, retrying", "backend", o.name, "status", resp.StatusCode, "err", errMsg, "attempt", attempt+1, "backoff", sleepDuration)
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(sleepDuration):
					backoff *= 2
					continue
				}
			}

			errMsg := "unknown error"
			if chatResp.Error != nil && chatResp.Error.Message != "" {
				errMsg = chatResp.Error.Message
			}
			return "", fmt.Errorf("%s API error (status %d): %s", o.name, resp.StatusCode, errMsg)
		}

		if !success {
			return "", fmt.Errorf("%s request failed after retries", o.name)
		}

		if len(chatResp.Choices) == 0 {
			return "", fmt.Errorf("%s returned no choices. body: %s", o.name, string(respBody))
		}

		results = append(results, chatResp.Choices[0].Message.Content)
	}

	return strings.Join(results, "\n\n"), nil
}

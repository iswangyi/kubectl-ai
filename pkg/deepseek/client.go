package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiEndpoint = "https://api.deepseek.com/v1/chat/completions" // 请根据实际的 DeepSeek API 端点调整
)

// Client 代表 DeepSeek API 客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient 创建新的 DeepSeek 客户端
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// TranslateCommand 将自然语言转换为 kubectl 命令
func (c *Client) TranslateCommand(ctx context.Context, naturalCommand string) (string, error) {
	prompt := fmt.Sprintf(`你是一个 Kubernetes 专家，请将以下自然语言描述转换为对应的 kubectl 命令。
只需要返回具体的命令，不需要任何解释。如果命令可能有危险性，请在命令前添加 [DANGEROUS] 标记。

自然语言描述: %s`, naturalCommand)

	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个 Kubernetes 专家，专门将自然语言转换为 kubectl 命令。",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	request := ChatRequest{
		Model:     "deepseek-chat", // 使用适当的模型名称
		Messages:  messages,
		MaxTokens: 150,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	command := response.Choices[0].Message.Content
	return command, nil
}

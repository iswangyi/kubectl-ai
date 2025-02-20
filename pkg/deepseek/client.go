package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	// deepseek
	apiEndpoint = "https://api.deepseek.com/chat/completions" // 请根据实际的 DeepSeek API 端点调整
	//阿里云
	//apiEndpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions" // 请根据实际的 DeepSeek API 端点调整
)

// Client 代表 DeepSeek API 客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
	messages   []Message
	enableChat bool
}

// NewClient 创建新的 DeepSeek 客户端
func NewClient(apiKey string, enableChat bool) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		messages:   make([]Message, 0),
		enableChat: enableChat,
	}
}

// sendChatRequest 发送聊天请求到 DeepSeek API
func (c *Client) sendChatRequest(ctx context.Context, newMessages []Message, stream bool) (string, error) {
    var messages []Message
    if c.enableChat {
        messages = append(c.messages, newMessages...)
        c.messages = messages
    } else {
        messages = newMessages
    }

    request := ChatRequest{
        Model:    "deepseek-chat",
        Messages: messages,
        Steam:    stream,
    }

    requestBody, err := json.Marshal(request)
    if err != nil {
        return "", fmt.Errorf("failed to marshal request: %v", err)
    }

    // 添加调试日志
    logrus.WithFields(logrus.Fields{
        "request_body": string(requestBody),
        "stream":      stream,
    }).Debug("Sending request to DeepSeek API")

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

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        // 添加错误响应的调试日志
        logrus.WithFields(logrus.Fields{
            "status_code": resp.StatusCode,
            "response":    string(body),
        }).Debug("API request failed")
        return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
    }

    var result strings.Builder
    if stream {
        reader := bufio.NewReader(resp.Body)

        for {
            line, err := reader.ReadString('\n')
            if err != nil {
                if err == io.EOF {
                    break
                }
                return "", fmt.Errorf("failed to read stream: %v", err)
            }

            line = strings.TrimSpace(line)
            if line == "" || line == "data: [DONE]" {
                continue
            }

            if !strings.HasPrefix(line, "data: ") {
                continue
            }

            data := strings.TrimPrefix(line, "data: ")
            var streamResp ChatStreamResponse
            if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
                return "", fmt.Errorf("failed to unmarshal stream response: %v", err)
            }

            if len(streamResp.Choices) > 0 {
                content := streamResp.Choices[0].Delta.Content
                fmt.Print(content)
                result.WriteString(content)
            }
        }

        return result.String(), nil
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read response: %v", err)
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

// TranslateCommand 将自然语言转换为 kubectl 命令
func (c *Client) TranslateCommand(ctx context.Context, naturalCommand string) (string, error) {
    prompt := fmt.Sprintf(`你是一个 Kubernetes 专家，请根据以下自然语言描述生成相应的 kubectl 命令。
请严格按照以下格式返回命令：
1. 如果需要获取集群信息，返回 [INFO] kubectl 命令
2. 如果是危险操作，返回 [DANGEROUS] kubectl 命令
3. 如果是普通操作，直接返回 kubectl 命令

禁止返回任何描述性文本，只返回实际可执行的命令,不明确的请不要用变量代替，后续我会在上下文中回传给你。

自然语言描述: %s`, naturalCommand)

    // 创建系统消息
    systemMessage := Message{
        Role:    "system",
        Content: "你是一个 Kubernetes 专家，专门将自然语言转换为 kubectl 命令。你需要先收集必要信息，再生成精确的执行命令。",
    }

    // 创建用户消息
    userMessage := Message{
        Role:    "user",
        Content: prompt,
    }

    // 如果启用了多轮对话，将新消息添加到历史消息中
    var messages []Message
    if c.enableChat {
        // 确保历史消息不会无限增长
        if len(c.messages) > 10 {
            // 保留最近的10条消息
            c.messages = c.messages[len(c.messages)-10:]
        }
        messages = append([]Message{systemMessage}, c.messages...)
        messages = append(messages, userMessage)
    } else {
        messages = []Message{systemMessage, userMessage}
    }

    // 发送请求并获取响应
    response, err := c.sendChatRequest(ctx, messages, false)
    if err != nil {
        return "", err
    }

    // 如果启用了多轮对话，保存用户消息和AI响应
    if c.enableChat {
        c.messages = append(c.messages, userMessage)
        c.messages = append(c.messages, Message{
            Role:    "assistant",
            Content: response,
        })
    }

    return response, nil
}

// ExplainCommand 解释kuberne中yaml、api-resources等的含义
func (c *Client) ExplainCommand(ctx context.Context, naturalCommand string) (string, error) {
    prompt := fmt.Sprintf("你是一个 Kubernetes 专家，请解释以下命令的含义。\n\n命令: %s", naturalCommand)
    messages := []Message{
        {
            Role:    "system",
            Content: "你是一个 Kubernetes 专家，专门解释命令的含义。",
        },
        {
            Role:    "user",
            Content: prompt,
        },
    }

    return c.sendChatRequest(ctx, messages,true)
}

// Message 表示对话消息
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// ChatRequest 表示发送到 DeepSeek API 的请求
type ChatRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Steam    bool      `json:"stream"`
}

// ChatResponse 表示 DeepSeek API 的非流式响应
type ChatResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    Choices []struct {
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
}

// ChatStreamResponse 表示 DeepSeek API 的流式响应
type ChatStreamResponse struct {
    ID      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    Choices []struct {
        Delta struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"delta"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
}


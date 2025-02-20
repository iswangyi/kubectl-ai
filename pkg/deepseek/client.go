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

	"github.com/yourusername/kubectl-ai/pkg/config"
)

const (
	// deepseek
	apiEndpoint = "https://api.deepseek.com/chat/completions" // 请根据实际的 DeepSeek API 端点调整
	//阿里云
	//apiEndpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions" // 请根据实际的 DeepSeek API 端点调整
)

// 全局消息历史
var globalMessages []Message

// Client 代表 DeepSeek API 客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
	enableChat bool
}

// NewClient 创建新的 DeepSeek 客户端
func NewClient(apiKey string, enableChat bool) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		enableChat: enableChat,
	}
}

// sendChatRequest 发送聊天请求到 DeepSeek API
func (c *Client) sendChatRequest(ctx context.Context, newMessages []Message, stream bool) (string, error) {
    var messages []Message
    if c.enableChat {
        // 优化消息合并逻辑，避免重复添加系统消息和用户消息
        messages = make([]Message, 0)
        lastUserContent := ""
        for _, msg := range globalMessages {
            if msg.Role == "user" {
                if msg.Content != lastUserContent {
                    messages = append(messages, msg)
                    lastUserContent = msg.Content
                }
            } else {
                messages = append(messages, msg)
            }
        }
        
        // 添加新消息，确保不重复
        for _, msg := range newMessages {
            if msg.Role == "system" {
                // 检查是否已存在系统消息
                hasSystem := false
                for _, existing := range messages {
                    if existing.Role == "system" {
                        hasSystem = true
                        break
                    }
                }
                if !hasSystem {
                    messages = append(messages, msg)
                }
            } else if msg.Role == "user" {
                // 检查是否与最后一条用户消息重复
                if msg.Content != lastUserContent {
                    messages = append(messages, msg)
                    lastUserContent = msg.Content
                }
            } else {
                messages = append(messages, msg)
            }
        }
        globalMessages = messages
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
    config.Logger.WithFields(map[string]interface{}{
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
        config.Logger.WithFields(map[string]interface{}{
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
	config.Logger.Debug(string(body))
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
    // 创建系统消息
    systemMessage := Message{
        Role:    "system",
        Content: `你是一个 Kubernetes 专家，专门将自然语言转换为 kubectl 命令。你需要先收集必要信息，再生成精确的执行命令。

请根据以下规则生成命令：
1. 获取集群信息 -> [INFO] kubectl 命令
2. 危险操作 -> [DANGEROUS] kubectl 命令
3. 普通操作 -> kubectl 命令
4. 禁止返回任何描述性文本，只返回实际可执行的命令
5. 不确定的不要用变量代替，后续会在上下文中补充`,
    }

    // 创建用户消息
    prompt := fmt.Sprintf("请将以下自然语言转换为 kubectl 命令：%s", naturalCommand)

    userMessage := Message{
        Role:    "user",
        Content: prompt,
    }

    var messages []Message
    if c.enableChat {
        // 确保历史消息不会无限增长
        if len(globalMessages) > 10 {
            globalMessages = globalMessages[len(globalMessages)-10:]
        }
        
        // 去重系统消息
        var hasSystemMessage bool
        for _, msg := range globalMessages {
            if msg.Role == "system" {
                hasSystemMessage = true
                break
            }
        }
        
        if !hasSystemMessage {
            messages = append(messages, systemMessage)
        }
        
        // 添加历史消息和新的用户消息
        messages = append(messages, globalMessages...)
        messages = append(messages, userMessage)
    } else {
        messages = []Message{systemMessage, userMessage}
    }

    // 发送请求并获取响应
    config.Logger.WithFields(map[string]interface{}{
        "messages_count": len(messages),
        "enable_chat":   c.enableChat,
    }).Debug("Sending messages to DeepSeek API")
    config.Logger.Debug(messages)

    response, err := c.sendChatRequest(ctx, messages, false)
    if err != nil {
        return "", err
    }

    // 如果启用了多轮对话，保存用户消息和AI响应
    if c.enableChat {
        globalMessages = append(globalMessages, userMessage)
        globalMessages = append(globalMessages, Message{
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


package kubectl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"github.com/yourusername/kubectl-ai/pkg/utils"
)

// Executor 代表 kubectl 命令执行器
type Executor struct {
    autoExecute bool
}

// NewExecutor 创建新的 kubectl 执行器
func NewExecutor(autoExecute bool) *Executor {
    return &Executor{
        autoExecute: autoExecute,
    }
}

// ExecuteNaturalCommand 执行自然语言转换后的 kubectl 命令
func (e *Executor) ExecuteNaturalCommand(ctx context.Context, kubectlCommand string) (string, error) {
    // 预处理 AI 返回的内容，提取实际命令
    lines := strings.Split(kubectlCommand, "\n")
    var commands []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        // 如果行以中文冒号结尾，说明是描述性文本，跳过
        if strings.HasSuffix(line, "：") {
            continue
        }
        // 如果行包含中文，说明是描述性文本，跳过
        if containsChinese(line) {
            continue
        }
        commands = append(commands, line)
    }

    // 如果没有提取到有效命令，返回错误
    if len(commands) == 0 {
        return "", fmt.Errorf("未能从 AI 响应中提取出有效的 kubectl 命令")
    }

    var lastOutput string
    for _, cmd := range commands {
        // 解析命令类型和实际命令
        cmdType, actualCmd := parseCommand(cmd)

        // 判断是否需要用户确认
        needConfirm := false
        var warningMsg string

        // 非查询命令或危险命令需要确认
        if !e.isQueryCommand(actualCmd) && !e.autoExecute {
            needConfirm = true
            if cmdType == "DANGEROUS" {
                warningMsg = fmt.Sprintf("\n%s即将执行危险命令：%s\n", utils.Yellow("[警告] "), actualCmd)
            } else {
                warningMsg = fmt.Sprintf("\n%s即将执行非查询命令：%s\n", utils.Yellow("[警告] "), actualCmd)
            }
        }

        // 如果需要确认，显示警告并获取用户确认
        if needConfirm {
            fmt.Print(warningMsg)
            if !confirmExecution() {
                return "", fmt.Errorf("用户取消了命令执行")
            }
        }

        // 根据命令类型执行不同的操作
        switch cmdType {
        case "INFO":
            // 执行信息收集命令
            output, err := e.executeCommand(ctx, actualCmd)
            if err != nil {
                return "", fmt.Errorf("执行信息收集命令失败: %v", err)
            }
            fmt.Printf("\n%s收集到的信息：%s\n", utils.Green("[INFO] "), output)
            lastOutput = output

        default:
            // 执行普通命令或危险命令
            fmt.Printf("\n%s执行命令：%s\n", utils.Blue("[执行] "), actualCmd)
            output, err := e.executeCommand(ctx, actualCmd)
            if err != nil {
                return "", fmt.Errorf("命令执行失败: %v", err)
            }
            lastOutput = output
        }
    }

    return lastOutput, nil
}

// containsChinese 检查字符串是否包含中文字符
func containsChinese(str string) bool {
    for _, r := range str {
        if r >= '\u4e00' && r <= '\u9fff' {
            return true
        }
    }
    return false
}

// parseCommand 解析命令类型和实际命令
func parseCommand(cmd string) (cmdType string, actualCmd string) {
    cmd = strings.TrimSpace(cmd)
    if strings.HasPrefix(cmd, "[INFO] ") {
        return "INFO", strings.TrimPrefix(cmd, "[INFO] ")
    }
    if strings.HasPrefix(cmd, "[DANGEROUS] ") {
        return "DANGEROUS", strings.TrimPrefix(cmd, "[DANGEROUS] ")
    }
    return "NORMAL", cmd
}

// executeCommand 执行 kubectl 命令
func (e *Executor) executeCommand(ctx context.Context, command string) (string, error) {
    args := strings.Fields(command)
    cmd := exec.CommandContext(ctx, "kubectl", args[1:]...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        // 如果命令执行失败，将错误输出和错误信息一起返回
        return string(output), fmt.Errorf("%v\n%s", err, output)
    }
    return string(output), err
}

// confirmExecution 询问用户是否确认执行命令
func confirmExecution() bool {
    fmt.Print("是否确认执行此命令？(y/n): ")
    var response string
    fmt.Scanln(&response)
    return strings.ToLower(response) == "y"
}

// isQueryCommand 判断是否为查询命令
func (e *Executor) isQueryCommand(kubectlCommand string) bool {
	// 标准化命令
	command := strings.TrimSpace(kubectlCommand)
	command = strings.TrimPrefix(command, "kubectl ")
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}
	action := parts[0]

	// 查询类命令白名单
	queryWhitelist := map[string]bool{
		"get": true, "describe": true, "explain": true, "logs": true,
		"top": true, "cluster-info": true, "attach": true, "exec": true,
		"proxy": true, "cp": true, "auth": true, "debug": true, "events": true,
	}

	// 高危写操作黑名单
	writeBlacklist := map[string]bool{
		"apply": true, "create": true, "delete": true, "patch": true,
		"replace": true, "scale": true, "rollout": true, "taint": true,
		"label": true, "annotate": true, "edit": true, "set": true,
	}

	// 混合模式判断
	if queryWhitelist[action] {
		return true
	}
	if writeBlacklist[action] {
		return false
	}

	// 默认保守策略：未知命令视为非查询
	return false
}




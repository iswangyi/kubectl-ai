package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/kubectl-ai/pkg/config"
	"github.com/yourusername/kubectl-ai/pkg/deepseek"
	"github.com/yourusername/kubectl-ai/pkg/kubectl"
	"github.com/yourusername/kubectl-ai/pkg/utils"
)


func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("Usage: kubectl ai \"<natural language command>\"")
		os.Exit(1)
	}

	// 获取自然语言命令
	naturalCommand := strings.Join(os.Args[1:], " ")

	// 创建上下文
	ctx := context.Background()

	// 创建 DeepSeek 客户端
	client := deepseek.NewClient(cfg.DeepseekAPIKey)

	// 创建 kubectl 执行器
	executor := kubectl.NewExecutor()

	// 调用 DeepSeek API 转换命令
	kubectlCommand, err := client.TranslateCommand(ctx, naturalCommand)
	if err != nil {
		fmt.Printf("Error translating command: %v\n", err)
		os.Exit(1)
	}

	// 检查是否是危险命令
	if strings.HasPrefix(kubectlCommand, "[DANGEROUS]") {
		kubectlCommand = strings.TrimPrefix(kubectlCommand, "[DANGEROUS]")
		fmt.Print(utils.FormatWarning("Warning: This command is potentially dangerous:", kubectlCommand))
		fmt.Print("\nDo you want to proceed? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Command aborted")
			os.Exit(0)
		}
	} else {
		// 非危险命令也使用红色显示
		fmt.Println(utils.FormatCommand(kubectlCommand))
	}


	// 执行 kubectl 命令
	if err := executor.Execute(ctx, kubectlCommand); err != nil {
		fmt.Printf("Error executing kubectl command: %v\n", err)
		os.Exit(1)
	}
}

// translateToKubectlCommand 使用 DeepSeek API 将自然语言转换为 kubectl 命令
func translateToKubectlCommand(ctx context.Context, naturalCommand string) (string, error) {
	// TODO: 实现 DeepSeek API 调用
	// 这里需要添加与 DeepSeek API 的集成代码
	return "", fmt.Errorf("not implemented yet")
}

// executeKubectlCommand 执行 kubectl 命令
func executeKubectlCommand(ctx context.Context, command string) error {
	// TODO: 实现 kubectl 命令执行
	// 这里需要添加执行 kubectl 命令的代码
	return fmt.Errorf("not implemented yet")
}

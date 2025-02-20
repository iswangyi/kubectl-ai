package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/kubectl-ai/pkg/config"
	"github.com/yourusername/kubectl-ai/pkg/deepseek"
	"github.com/yourusername/kubectl-ai/pkg/kubectl"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// 检查命令行参数
	if len(os.Args) < 3 {
		fmt.Println("Usage: kubectl ai <cmd|explain> \"<natural language command>\"")
		os.Exit(1)
	}

	// 获取子命令和自然语言命令
	subCommand := os.Args[1]
	naturalCommand := strings.Join(os.Args[2:], " ")

	// 创建上下文
	ctx := context.Background()

	// 创建 DeepSeek 客户端
	client := deepseek.NewClient(cfg.DeepseekAPIKey, cfg.EnableChat)

	// 创建 kubectl 执行器
	executor := kubectl.NewExecutor(cfg.AutoExecute)

	// 根据子命令执行不同的操作
	switch subCommand {
	case "cmd":
		// 调用 DeepSeek API 转换命令
		kubectlCommand, err := client.TranslateCommand(ctx, naturalCommand)
		if err != nil {
			fmt.Printf("Error translating command: %v\n", err)
			os.Exit(1)
		}

		// 执行命令并获取输出
		output, err := executor.ExecuteNaturalCommand(ctx, kubectlCommand)
		if err != nil {
			fmt.Printf("Error executing command: %v\n", err)
			os.Exit(1)
		}

		// 如果命令执行有输出，询问用户是否继续与 AI 交互
		if output != "" {
			fmt.Print("是否将执行结果传递给 AI 继续处理？(y/n): ")
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) == "y" {
				naturalCommand = fmt.Sprintf("%s\n命令执行结果：\n%s", naturalCommand, output)
				kubectlCommand, err = client.TranslateCommand(ctx, naturalCommand)
				if err != nil {
					fmt.Printf("Error translating command with context: %v\n", err)
					os.Exit(1)
				}

				// 执行新的命令
				if _, err := executor.ExecuteNaturalCommand(ctx, kubectlCommand); err != nil {
					fmt.Printf("Error executing command: %v\n", err)
					os.Exit(1)
				}
			}
		}

	case "explain":
		// 调用 DeepSeek API 解释命令
		explanation, err := client.ExplainCommand(ctx, naturalCommand)
		if err != nil {
			fmt.Printf("Error explaining command: %v\n", err)
			os.Exit(1)
		}

		// 打印解释
		fmt.Println(explanation)

	default:
		fmt.Printf("Unknown subcommand: %s\n", subCommand)
		fmt.Println("Usage: kubectl ai <cmd|explain> \"<natural language command>\"")
		os.Exit(1)
	}
}

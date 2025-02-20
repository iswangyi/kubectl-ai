package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"bufio"

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

		// 如果有输出，直接打印
		if output != "" {
			fmt.Printf("\n%s\n", output)
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
	case "exec":
		// 进入交互模式
		fmt.Println("进入交互模式，输入 'exit' 退出")
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("\n请输入问题:")
			var input string
			if scanner.Scan() {
				input = scanner.Text()
			}

			if input == "exit" {
				fmt.Println("退出交互模式")
				break
			}

			// 调用 DeepSeek API 转换命令
			kubectlCommand, err := client.TranslateCommand(ctx, input)
			if err != nil {
				fmt.Printf("Error translating command: %v\n", err)
				continue
			}

			// 执行命令并获取输出
			output, err := executor.ExecuteNaturalCommand(ctx, kubectlCommand)
			if err != nil {
				fmt.Printf("Error executing command: %v\n", err)
				continue
			}

			// 如果有输出，直接打印
			if output != "" {
				fmt.Printf("\n%s\n", output)
			}

			// 询问用户是否继续
			for {
				fmt.Print("\n是否基于当前结果继续对话？(y/n): ")
				var response string
				if scanner.Scan() {
					response = scanner.Text()
				}
				response = strings.ToLower(strings.TrimSpace(response))
				if response == "y" || response == "n" {
					if response == "y" {
						// 将当前输出作为上下文
						fmt.Print("\n请输入新的问题: ")
						if scanner.Scan() {
							input = scanner.Text()
							// 将新问题和上下文一起提交给 AI
							contextCommand := fmt.Sprintf("基于上次执行结果：%s\n新的问题：%s", output, input)
							kubectlCommand, err = client.TranslateCommand(ctx, contextCommand)
							if err != nil {
								fmt.Printf("Error translating command with context: %v\n", err)
								break
							}

							// 执行新的命令
							output, err = executor.ExecuteNaturalCommand(ctx, kubectlCommand)
							if err != nil {
								fmt.Printf("Error executing command: %v\n", err)
							}
							// 如果有输出，直接打印
							if output != "" {
								fmt.Printf("\n%s\n", output)
							}
						}
					}
					break
				}
				fmt.Println("请输入 y 或 n")
			}
		}

	default:
		fmt.Printf("Unknown subcommand: %s\n", subCommand)
		fmt.Println("Usage: kubectl ai <cmd|explain|exec> \"<natural language command>\"")
		os.Exit(1)
	}
}

package kubectl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Executor 处理 kubectl 命令的执行
type Executor struct {}

// NewExecutor 创建新的 kubectl 执行器
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute 执行 kubectl 命令
func (e *Executor) Execute(ctx context.Context, command string) error {
	// 获取当前的 kubeconfig 路径
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		// 如果环境变量未设置，使用默认路径
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %v", err)
		}
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	// 将命令分割为参数数组
	args := strings.Fields(command)
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	// 创建命令并设置环境变量
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("KUBECONFIG=%s", kubeconfig),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)

	// 执行命令
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute kubectl command: %v", err)
	}

	return nil
}


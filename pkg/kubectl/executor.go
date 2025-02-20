package kubectl

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Executor 处理 kubectl 命令的执行
type Executor struct{}

// NewExecutor 创建新的 kubectl 执行器
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute 执行 kubectl 命令
func (e *Executor) Execute(ctx context.Context, command string) error {
	args := strings.Fields(command)
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute kubectl command: %v, output: %s", err, string(output))
	}

	fmt.Println(string(output))
	return nil
}

# kubectl-ai

[![Go Version](https://img.shields.io/github/go-mod/go-version/yourusername/kubectl-ai)](https://github.com/yourusername/kubectl-ai)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

kubectl-ai 是一个基于自然语言处理的 kubectl 命令行工具，它能够将自然语言转换为 kubectl 命令，并提供命令解释功能。通过集成 DeepSeek API，该工具让 Kubernetes 集群管理变得更加简单和直观。

## 特性

- 🤖 自然语言转换为 kubectl 命令
- 📝 命令解释功能
- 💬 支持多轮对话
- ⚡ 自动执行模式
- 🔒 安全性检查
- 🌈 彩色输出

## 安装

### 前置条件

- Go 1.21 或更高版本
- kubectl 命令行工具
- DeepSeek API 密钥

### 从源码安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/kubectl-ai.git
cd kubectl-ai

# 编译安装
make build
make install
```

## 配置

1. 创建配置文件 `config.yaml`：

```yaml
# DeepSeek API 配置
deepseek:
  api_key: "your-api-key-here" # 可通过环境变量 DEEPSEEK_API_KEY 覆盖

# 执行配置
auto_execute: false # 可通过环境变量 AUTO_EXECUTE 覆盖

# 聊天配置
enable_chat: true # 可通过环境变量 ENABLE_CHAT 覆盖

# 调试配置
LOG_LEVEL: debug # 可通过环境变量 DEBUG 覆盖
```

2. 设置环境变量（可选）：

```bash
export DEEPSEEK_API_KEY="your-api-key-here"
export AUTO_EXECUTE=true
export ENABLE_CHAT=true
export DEBUG=true
```

## 使用方法

### 命令转换模式

```bash
kubectl ai cmd "显示所有命名空间的 pod"
```

### 命令解释模式

```bash
kubectl ai explain "kubectl get pods -n kube-system"
```

### 交互模式

```bash
kubectl ai exec
```

## 示例

1. 查询 Pod 状态：
```bash
kubectl ai cmd "查看所有命名空间的 pod 状态"
```

2. 解释命令：
```bash
kubectl ai explain "kubectl describe pod nginx-pod -n default"
```

3. 交互式操作：
```bash
kubectl ai exec
> 查看所有命名空间的 deployment
> 显示 kube-system 命名空间下的所有服务
> exit
```

## 安全性

- 危险操作会显示警告并要求确认
- 查询操作无需确认
- 支持命令白名单和黑名单

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

[MIT License](LICENSE)
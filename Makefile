# 获取当前的操作系统和 CPU 架构
SHELL=/bin/bash
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GOOS := linxu
GOARCH := x86 

# 定义构建目标
BUILD_TARGET := kubectl-ai

# 定义构建的源文件路径
BUILD_SRC := cmd/kubectl-ai/main.go

# 默认目标
all: build install

# 构建目标
build:
	@echo "Building $(BUILD_TARGET) for $(GOOS)/$(GOARCH)..."
	go build -o $(BUILD_TARGET) $(BUILD_SRC)
	@echo "Build completed."

# 添加可执行权限
chmod:
	chmod +x $(BUILD_TARGET)

# 安装目标
install: build chmod
	@echo "Installing $(BUILD_TARGET) to /usr/local/bin..."
	mv $(BUILD_TARGET) /usr/local/bin
	@echo "Installation completed."

# 清理目标
clean:
	@echo "Cleaning up..."
	rm -f $(BUILD_TARGET)
	@echo "Cleanup completed."

# 帮助信息
help:
	@echo "Usage: make [target]"
	@echo "Targets:"
	@echo "  all       Build and install $(BUILD_TARGET)"
	@echo "  build     Build $(BUILD_TARGET)"
	@echo "  chmod     Add executable permission to $(BUILD_TARGET)"
	@echo "  install   Install $(BUILD_TARGET) to /usr/local/bin"
	@echo "  clean     Clean up the built binary"
	@echo "  help      Show this help message"

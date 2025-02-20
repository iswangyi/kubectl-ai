package utils

import "fmt"

// Blue 返回蓝色文本
func Blue(text string) string {
	return fmt.Sprintf("\x1b[34m%s\x1b[0m", text)
}

// Green 返回绿色文本
func Green(text string) string {
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", text)
}

// Yellow 返回黄色文本
func Yellow(text string) string {
	return fmt.Sprintf("\x1b[33m%s\x1b[0m", text)
}

// Red 返回红色文本
func Red(text string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", text)
}

// FormatCommand 格式化命令显示
func FormatCommand(command string) string {
	return Red(command)
}

// FormatWarning 格式化警告信息
func FormatWarning(warning string, command string) string {
	return fmt.Sprintf("%s\n%s", Red(warning), Red(command))
}
package utils

import "fmt"

// ColorText 为文本添加颜色
type ColorText struct {
	Red    func(text string) string
	Yellow func(text string) string
	Green  func(text string) string
}

// NewColorText 创建新的颜色文本处理器
func NewColorText() *ColorText {
	return &ColorText{
		Red: func(text string) string {
			return fmt.Sprintf("\x1b[31m%s\x1b[0m", text)
		},
		Yellow: func(text string) string {
			return fmt.Sprintf("\x1b[33m%s\x1b[0m", text)
		},
		Green: func(text string) string {
			return fmt.Sprintf("\x1b[32m%s\x1b[0m", text)
		},
	}
}

// FormatCommand 格式化命令显示
func FormatCommand(command string) string {
	color := NewColorText()
	return color.Red(command)
}

// FormatWarning 格式化警告信息
func FormatWarning(warning string, command string) string {
	color := NewColorText()
	return fmt.Sprintf("%s\n%s", color.Red(warning), color.Red(command))
}
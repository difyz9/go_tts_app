package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"tts_app/service"
)

func main() {
	// 创建文本处理器
	processor := service.NewTextProcessor()
	
	// 读取测试文件
	file, err := os.Open("test_markdown.md")
	if err != nil {
		fmt.Printf("无法打开测试文件: %v\n", err)
		return
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	
	fmt.Println("=== Markdown文本处理测试结果 ===\n")
	
	for scanner.Scan() {
		lineNumber++
		originalLine := scanner.Text()
		
		// 检查是否为有效TTS文本
		isValid := processor.IsValidTextForTTS(originalLine)
		
		if !isValid {
			fmt.Printf("行 %d [跳过]: %s\n", lineNumber, originalLine)
			continue
		}
		
		// 处理文本
		processedText := processor.ProcessText(originalLine)
		
		if strings.TrimSpace(processedText) == "" {
			fmt.Printf("行 %d [处理后为空]: %s\n", lineNumber, originalLine)
			continue
		}
		
		// 显示处理结果
		if originalLine != processedText {
			fmt.Printf("行 %d [已处理]:\n", lineNumber)
			fmt.Printf("  原文: %s\n", originalLine)
			fmt.Printf("  处理后: %s\n", processedText)
		} else {
			fmt.Printf("行 %d [保留]: %s\n", lineNumber, originalLine)
		}
		fmt.Println()
	}
	
	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件时出错: %v\n", err)
	}
}

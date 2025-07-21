package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"tts_app/service"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run test_text_processor.go <输入文件>")
		return
	}

	inputFile := os.Args[1]
	
	// 创建文本处理器
	textProcessor := service.NewTextProcessor()
	
	// 读取输入文件
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Println("=== TTS文本处理测试 ===")
	fmt.Println()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	validCount := 0
	skippedCount := 0

	for scanner.Scan() {
		originalText := scanner.Text()
		
		fmt.Printf("第%d行原文: %s\n", lineNum, originalText)
		
		// 快速过滤逻辑（与服务中的逻辑一致）
		trimmedLine := strings.TrimSpace(originalText)
		if trimmedLine == "" {
			fmt.Printf("  -> 跳过: 空行\n")
			skippedCount++
		} else if strings.HasPrefix(trimmedLine, "## ") ||
			strings.HasPrefix(trimmedLine, "### ") ||
			strings.HasPrefix(trimmedLine, "#### ") ||
			strings.HasPrefix(trimmedLine, "** ") ||
			strings.HasPrefix(trimmedLine, "| ") ||
			trimmedLine == "##" ||
			trimmedLine == "###" ||
			trimmedLine == "####" ||
			trimmedLine == "**" ||
			trimmedLine == "***" ||
			strings.HasPrefix(trimmedLine, "-- ") ||
			strings.HasPrefix(trimmedLine, "-----") {
			fmt.Printf("  -> 跳过: 快速过滤（标记行）\n")
			skippedCount++
		} else if !textProcessor.IsValidTextForTTS(originalText) {
			fmt.Printf("  -> 跳过: 详细验证（无效文本行）\n")
			skippedCount++
		} else {
			// 处理文本
			processedText := textProcessor.ProcessText(originalText)
			if processedText == "" {
				fmt.Printf("  -> 跳过: 处理后为空\n")
				skippedCount++
			} else {
				fmt.Printf("  -> 处理后: %s\n", processedText)
				validCount++
			}
		}
		
		fmt.Println()
		lineNum++
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	fmt.Printf("=== 处理统计 ===\n")
	fmt.Printf("总行数: %d\n", lineNum-1)
	fmt.Printf("有效行数: %d\n", validCount)
	fmt.Printf("跳过行数: %d\n", skippedCount)
}

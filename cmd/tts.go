/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/difyz9/markdown2tts/service"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var configFile string
var inputFile string
var outputDir string
var ttsSmartMarkdown bool // 新增：智能Markdown模式

// ttsCmd represents the tts command
var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "将文本文件转换为语音并合并（默认并发模式）",
	Long: `使用腾讯云TTS服务将文本文件转换为语音，并自动合并成一个音频文件。

默认启用并发处理模式，自动加载配置文件，操作简单快捷。
当输入文件为Markdown格式（.md或.markdown）时，自动启用智能Markdown处理模式。

示例:
  markdown2tts tts                                    # 使用默认配置
  markdown2tts tts -i input.txt                       # 指定输入文件
  markdown2tts tts -i document.md                     # 自动启用智能Markdown模式
  markdown2tts tts -i input.txt -o /path/to/output   # 指定输入和输出
  markdown2tts tts --config custom.yaml              # 使用自定义配置
  `,
	Run: func(cmd *cobra.Command, args []string) {
		err := runTTS(cmd)
		if err != nil {
			fmt.Printf("错误: %v\n", err)
		}
	},
}

func runTTS(cmd *cobra.Command) error {
	// 如果没有指定配置文件，尝试默认位置
	if configFile == "" {
		configFile = "config.yaml"
	}

	// 加载配置（如果配置文件不存在会自动初始化）
	configService, err := service.NewConfigService(configFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	config := configService.GetConfig()

	// 如果指定了输入文件，覆盖配置
	if inputFile != "" {
		config.InputFile = inputFile

		// 自动检测markdown文件并启用智能处理模式（仅当用户未明确设置smart-markdown标志时）
		ext := strings.ToLower(filepath.Ext(inputFile))
		if ext == ".md" || ext == ".markdown" {
			// 检查用户是否明确设置了smart-markdown标志
			smartMarkdownSet := cmd.Flags().Changed("smart-markdown")
			if !smartMarkdownSet {
				ttsSmartMarkdown = true
				fmt.Printf("🔍 检测到Markdown文件，自动启用智能Markdown处理模式\n")
			}
		}
	}

	// 如果指定了输出目录，覆盖配置
	if outputDir != "" {
		config.Audio.OutputDir = outputDir
	}

	// 验证配置
	if config.TencentCloud.SecretID == "your_secret_id" || config.TencentCloud.SecretKey == "your_secret_key" {
		return fmt.Errorf("请在配置文件中设置正确的腾讯云SecretID和SecretKey")
	}

	// 创建TTS服务
	ttsService := service.NewTTSService(
		config.TencentCloud.SecretID,
		config.TencentCloud.SecretKey,
		config.TencentCloud.Region,
		config,
	)

	if ttsService == nil {
		return fmt.Errorf("创建TTS服务失败")
	}

	// 检查输入文件路径
	historyPath := config.InputFile
	if !filepath.IsAbs(historyPath) {
		// 如果是相对路径，基于当前工作目录
		absPath, err := filepath.Abs(historyPath)
		if err != nil {
			return fmt.Errorf("无法解析输入文件路径: %v", err)
		}
		historyPath = absPath
		config.InputFile = historyPath
	}

	// 创建输出目录
	if err := service.EnsureDir(config.Audio.OutputDir); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	fmt.Printf("配置信息:\n")
	fmt.Printf("- 输入文件: %s\n", config.InputFile)
	fmt.Printf("- 音色: %d\n", config.TTS.VoiceType)
	fmt.Printf("- 语速: %.1f\n", config.TTS.Speed)
	fmt.Printf("- 音量: %d\n", config.TTS.Volume)
	fmt.Printf("- 输出目录: %s\n", config.Audio.OutputDir)
	fmt.Printf("- 最终文件: %s\n", config.Audio.FinalOutput)
	fmt.Printf("- 并发模式: 开启（默认）\n")
	fmt.Printf("- 最大并发数: %d\n", config.Concurrent.MaxWorkers)
	fmt.Printf("- 速率限制: %d次/秒\n", config.Concurrent.RateLimit)

	// 显示处理模式
	if ttsSmartMarkdown {
		fmt.Printf("- 处理模式: 智能Markdown模式（blackfriday解析）\n")
	} else {
		fmt.Printf("- 处理模式: 传统逐行模式\n")
	}
	fmt.Println()

	// 默认使用并发处理模式
	concurrentAudioService := service.NewConcurrentAudioService(config, ttsService)

	// 根据模式选择处理方法
	if ttsSmartMarkdown {
		fmt.Println("开始智能Markdown处理（腾讯云TTS）...")
		err = concurrentAudioService.ProcessMarkdownFileConcurrent()
	} else {
		fmt.Println("开始并发处理文本文件（腾讯云TTS）...")
		err = concurrentAudioService.ProcessInputFileConcurrent()
	}

	if err != nil {
		return fmt.Errorf("处理文件失败: %v", err)
	}

	fmt.Println("TTS转换和音频合并完成！")
	return nil
}

func init() {
	rootCmd.AddCommand(ttsCmd)

	// 添加配置文件标志（可选）
	ttsCmd.Flags().StringVarP(&configFile, "config", "c", "", "配置文件路径（默认自动查找config.yaml）")

	// 添加输入文件标志
	ttsCmd.Flags().StringVarP(&inputFile, "input", "i", "", "输入文本文件路径")

	// 添加输出目录标志
	ttsCmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出目录路径（默认为./output）")

	// 添加智能Markdown处理标志
	ttsCmd.Flags().BoolVar(&ttsSmartMarkdown, "smart-markdown", false, "启用智能Markdown处理模式（推荐用于.md文件）")
}

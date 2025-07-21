/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"path/filepath"
	"tts_app/service"

	"github.com/spf13/cobra"
)

var configFile string
var inputFile string
var outputDir string

// ttsCmd represents the tts command
var ttsCmd = &cobra.Command{
	Use:   "tts",
	Short: "将文本文件转换为语音并合并（默认并发模式）",
	Long: `使用腾讯云TTS服务将文本文件转换为语音，并自动合并成一个音频文件。

默认启用并发处理模式，自动加载配置文件，操作简单快捷。

示例:
  tts_app tts                                    # 使用默认配置
  tts_app tts -i input.txt                       # 指定输入文件
  tts_app tts -i input.txt -o /path/to/output   # 指定输入和输出
  tts_app tts --config custom.yaml              # 使用自定义配置
  # 智能Markdown模式（推荐用于.md文件）
  tts_app edge -i document.md --smart-markdown -o output
  # 传统模式（用于纯文本文件）
  tts_app edge -i document.txt -o output
  `,
	Run: func(cmd *cobra.Command, args []string) {
		err := runTTS()
		if err != nil {
			fmt.Printf("错误: %v\n", err)
		}
	},
}

func runTTS() error {
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
	fmt.Println()

	// 默认使用并发处理模式
	fmt.Println("开始并发处理文本文件...")
	concurrentAudioService := service.NewConcurrentAudioService(config, ttsService)
	err = concurrentAudioService.ProcessInputFileConcurrent()
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
}

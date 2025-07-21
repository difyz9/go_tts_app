/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"github.com/difyz9/markdown2tts/service"
	"path/filepath"

	"github.com/spf13/cobra"
)

var edgeConfigFile string
var edgeInputFile string
var edgeOutputDir string
var listVoices string
var listAllVoices bool
var edgeVoice string
var edgeRate string
var edgeVolume string
var edgePitch string
var smartMarkdown bool // 新增：智能Markdown模式

// edgeCmd represents the edge command
var edgeCmd = &cobra.Command{
	Use:   "edge",
	Short: "使用Edge TTS进行语音合成（默认并发模式）",
	Long: `使用Microsoft Edge TTS服务将文本文件转换为语音，并自动合并成一个音频文件。

默认启用并发处理模式，自动加载配置文件，操作简单快捷。
Edge TTS是免费的，无需API密钥，支持多种语言和音色。

示例:
  github.com/difyz9/markdown2tts edge                                    # 使用默认配置
  github.com/difyz9/markdown2tts edge -i input.txt                       # 指定输入文件
  github.com/difyz9/markdown2tts edge -i input.txt -o /path/to/output   # 指定输入和输出
  github.com/difyz9/markdown2tts edge --config custom.yaml              # 使用自定义配置
  github.com/difyz9/markdown2tts edge --list-all                         # 列出所有可用语音
  github.com/difyz9/markdown2tts edge --list zh                          # 列出中文语音
  github.com/difyz9/markdown2tts edge --list en                          # 列出英文语音
  github.com/difyz9/markdown2tts edge --voice zh-CN-YunyangNeural        # 使用指定语音
  github.com/difyz9/markdown2tts edge --rate +20% --volume +10%          # 调整语速和音量
  # 智能Markdown模式（推荐用于.md文件）
  github.com/difyz9/markdown2tts edge -i document.md --smart-markdown -o output
  # 传统模式（用于纯文本文件）
  github.com/difyz9/markdown2tts edge -i document.txt -o output
  `,
	Run: func(cmd *cobra.Command, args []string) {
		err := runEdgeTTS()
		if err != nil {
			fmt.Printf("错误: %v\n", err)
		}
	},
}

func runEdgeTTS() error {
	// 如果是列出语音模式，直接执行并返回
	if listAllVoices || listVoices != "" {
		if listAllVoices {
			return service.ListEdgeVoices("")
		}
		return service.ListEdgeVoices(listVoices)
	}

	// 如果没有指定配置文件，尝试默认位置
	if edgeConfigFile == "" {
		edgeConfigFile = "config.yaml"
	}

	// 加载配置（如果配置文件不存在会自动初始化）
	configService, err := service.NewConfigService(edgeConfigFile)
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	config := configService.GetConfig()

	// 如果指定了输入文件，覆盖配置
	if edgeInputFile != "" {
		config.InputFile = edgeInputFile
	}

	// 如果指定了输出目录，覆盖配置
	if edgeOutputDir != "" {
		config.Audio.OutputDir = edgeOutputDir
	}

	// 如果指定了语音参数，覆盖配置
	if edgeVoice != "" {
		config.EdgeTTS.Voice = edgeVoice
	}
	if edgeRate != "" {
		config.EdgeTTS.Rate = edgeRate
	}
	if edgeVolume != "" {
		config.EdgeTTS.Volume = edgeVolume
	}
	if edgePitch != "" {
		config.EdgeTTS.Pitch = edgePitch
	}

	// 检查输入文件路径
	inputPath := config.InputFile
	if !filepath.IsAbs(inputPath) {
		// 如果是相对路径，基于当前工作目录
		absPath, err := filepath.Abs(inputPath)
		if err != nil {
			return fmt.Errorf("无法解析输入文件路径: %v", err)
		}
		inputPath = absPath
		config.InputFile = inputPath
	}

	// 创建输出目录
	if err := service.EnsureDir(config.Audio.OutputDir); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	fmt.Printf("配置信息:\n")
	fmt.Printf("- 输入文件: %s\n", config.InputFile)
	fmt.Printf("- 输出目录: %s\n", config.Audio.OutputDir)
	fmt.Printf("- 最终文件: %s\n", config.Audio.FinalOutput)
	fmt.Printf("- 并发模式: 开启（默认）\n")
	fmt.Printf("- 最大并发数: %d\n", config.Concurrent.MaxWorkers)
	fmt.Printf("- 速率限制: %d次/秒\n", config.Concurrent.RateLimit)
	fmt.Printf("- TTS引擎: Microsoft Edge TTS (免费)\n")

	// 显示Edge TTS配置
	voice := config.EdgeTTS.Voice
	if voice == "" {
		voice = "zh-CN-XiaoyiNeural"
	}
	rate := config.EdgeTTS.Rate
	if rate == "" {
		rate = "+0%"
	}
	volume := config.EdgeTTS.Volume
	if volume == "" {
		volume = "+0%"
	}
	pitch := config.EdgeTTS.Pitch
	if pitch == "" {
		pitch = "+0Hz"
	}

	fmt.Printf("- 语音: %s\n", voice)
	fmt.Printf("- 语速: %s\n", rate)
	fmt.Printf("- 音量: %s\n", volume)
	fmt.Printf("- 音调: %s\n", pitch)

	// 显示处理模式
	if smartMarkdown {
		fmt.Printf("- 处理模式: 智能Markdown模式（blackfriday解析）\n")
	} else {
		fmt.Printf("- 处理模式: 传统逐行模式\n")
	}
	fmt.Println()

	// 创建Edge TTS服务
	edgeService := service.NewEdgeTTSService(config)

	// 根据模式选择处理方法
	if smartMarkdown {
		fmt.Println("开始智能Markdown处理（Edge TTS）...")
		err = edgeService.ProcessMarkdownFile(config.InputFile, config.Audio.OutputDir)
	} else {
		fmt.Println("开始并发处理文本文件（Edge TTS）...")
		err = edgeService.ProcessInputFileConcurrent()
	}

	if err != nil {
		return fmt.Errorf("处理文件失败: %v", err)
	}

	fmt.Println("Edge TTS转换和音频合并完成！")
	return nil
}

func init() {
	rootCmd.AddCommand(edgeCmd)

	// 添加配置文件标志（可选）
	edgeCmd.Flags().StringVarP(&edgeConfigFile, "config", "c", "", "配置文件路径（默认自动查找config.yaml）")

	// 添加输入文件标志
	edgeCmd.Flags().StringVarP(&edgeInputFile, "input", "i", "", "输入文本文件路径")

	// 添加输出目录标志
	edgeCmd.Flags().StringVarP(&edgeOutputDir, "output", "o", "", "输出目录路径（默认为./output）")

	// 添加列出语音标志
	edgeCmd.Flags().BoolVar(&listAllVoices, "list-all", false, "列出所有可用语音")
	edgeCmd.Flags().StringVar(&listVoices, "list", "", "列出指定语言的语音（如: zh, en, ja）")

	// 添加语音参数标志
	edgeCmd.Flags().StringVar(&edgeVoice, "voice", "", "指定语音 (如: zh-CN-XiaoyiNeural)")
	edgeCmd.Flags().StringVar(&edgeRate, "rate", "", "语速 (如: +20%, -10%)")
	edgeCmd.Flags().StringVar(&edgeVolume, "volume", "", "音量 (如: +10%, -20%)")
	edgeCmd.Flags().StringVar(&edgePitch, "pitch", "", "音调 (如: +10Hz, -5Hz)")

	// 添加智能Markdown处理标志
	edgeCmd.Flags().BoolVar(&smartMarkdown, "smart-markdown", false, "启用智能Markdown处理模式（推荐用于.md文件）")
}

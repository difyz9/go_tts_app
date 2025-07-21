/*
TTS语音合成应用 - 根命令定义

Copyright © 2025 TTS App Contributors
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 版本信息
var (
	appVersion   = "dev"
	appBuildTime = "unknown"
	appGitCommit = "unknown"
)

// SetVersionInfo 设置版本信息
func SetVersionInfo(version, buildTime, gitCommit string) {
	appVersion = version
	appBuildTime = buildTime
	appGitCommit = gitCommit

	// 更新rootCmd的版本信息
	rootCmd.Version = getVersionString()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "markdown2tts",
	Short: "🎵 TTS语音合成应用 - 支持双引擎、并发处理的高性能文本转语音工具",
	Long: `🎵 TTS语音合成应用

一个功能完整、高性能的文本转语音(TTS)应用程序，支持双引擎、并发处理、智能过滤等特色功能。

✨ 核心特色：
  🎯 双引擎支持    - 腾讯云TTS + Microsoft Edge TTS  
  🚀 并发处理      - 最高20倍速度提升
  🆓 完全免费      - Edge TTS无需API密钥
  🔧 智能过滤      - 自动跳过无效文本
  📊 实时进度      - 详细处理状态显示
  🌍 跨平台支持    - Windows/macOS/Linux

🚀 快速开始：
  # 初始化配置（新用户）
  markdown2tts init
  
  # 免费转换（推荐）
  markdown2tts edge -i input.txt
  
  # 企业用户
  markdown2tts tts -i input.txt
  
  # 查看语音选项  
  markdown2tts edge --list zh📚 更多信息：https://github.com/difyz9/markdown2tts`,
	Version: getVersionString(),
}

// getVersionString 获取版本字符串
func getVersionString() string {
	if appVersion == "dev" {
		return fmt.Sprintf("%s (commit: %s, built: %s)", appVersion, appGitCommit, appBuildTime)
	}
	return fmt.Sprintf("%s (commit: %s, built: %s)", appVersion, appGitCommit, appBuildTime)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// 设置版本模板
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`)

	// 全局标志
	rootCmd.PersistentFlags().BoolP("help", "h", false, "显示帮助信息")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "显示版本信息")

	// 设置帮助标志不显示在使用说明中
	rootCmd.PersistentFlags().MarkHidden("help")
}

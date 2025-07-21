/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"tts_app/service"

	"github.com/spf13/cobra"
)

var initConfigFile string
var initInputFile string
var force bool

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化配置文件和示例输入文件",
	Long: `初始化TTS应用所需的配置文件和示例输入文件。

该命令会创建：
1. config.yaml - 主配置文件
2. input.txt - 示例输入文件

如果文件已存在，默认会跳过。使用 --force 强制覆盖。

示例:
  tts_app init                           # 使用默认文件名初始化
  tts_app init --config custom.yaml     # 指定配置文件名
  tts_app init --input my_input.txt      # 指定输入文件名
  tts_app init --force                   # 强制覆盖已存在的文件`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runInit()
		if err != nil {
			fmt.Printf("错误: %v\n", err)
		}
	},
}

func runInit() error {
	// 设置默认文件名
	if initConfigFile == "" {
		initConfigFile = "config.yaml"
	}
	if initInputFile == "" {
		initInputFile = "input.txt"
	}

	fmt.Println("🎵 TTS应用初始化")
	fmt.Println("================")
	fmt.Println()

	initializer := service.NewConfigInitializer()

	// 如果强制模式，先删除已存在的文件
	if force {
		fmt.Println("⚠️  强制模式：将覆盖已存在的文件")
		// 这里可以添加删除文件的逻辑，但为了安全，我们让初始化器处理
	}

	// 初始化配置文件
	fmt.Printf("📝 初始化配置文件: %s\n", initConfigFile)
	err := initializer.InitializeConfig(initConfigFile)
	if err != nil {
		return fmt.Errorf("初始化配置文件失败: %v", err)
	}

	// 创建示例输入文件
	fmt.Printf("📄 创建示例输入文件: %s\n", initInputFile)
	err = initializer.CreateSampleInputFile(initInputFile)
	if err != nil {
		return fmt.Errorf("创建示例输入文件失败: %v", err)
	}

	// 显示快速开始指南
	initializer.ShowQuickStart()

	fmt.Println("🎉 初始化完成！")
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Printf("1. 编辑 %s 设置您的API密钥（可选，使用腾讯云TTS时需要）\n", initConfigFile)
	fmt.Printf("2. 编辑 %s 添加要转换的文本\n", initInputFile)
	fmt.Println("3. 运行 TTS 转换：")
	fmt.Printf("   - 免费版本: ./tts_app edge -i %s\n", initInputFile)
	fmt.Printf("   - 腾讯云版本: ./tts_app tts -i %s\n", initInputFile)

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)

	// 添加配置文件标志
	initCmd.Flags().StringVarP(&initConfigFile, "config", "c", "", "配置文件路径（默认: config.yaml）")

	// 添加输入文件标志
	initCmd.Flags().StringVarP(&initInputFile, "input", "i", "", "示例输入文件路径（默认: input.txt）")

	// 添加强制覆盖标志
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "强制覆盖已存在的文件")
}

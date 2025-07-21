package service

import (
	"fmt"
	"github.com/difyz9/markdown2tts/model"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigInitializer 配置初始化器
type ConfigInitializer struct{}

// NewConfigInitializer 创建配置初始化器
func NewConfigInitializer() *ConfigInitializer {
	return &ConfigInitializer{}
}

// InitializeConfig 初始化配置文件
func (ci *ConfigInitializer) InitializeConfig(configPath string) error {
	return ci.InitializeConfigWithForce(configPath, false)
}

// InitializeConfigWithForce 初始化配置文件（支持强制覆盖）
func (ci *ConfigInitializer) InitializeConfigWithForce(configPath string, force bool) error {
	// 检查配置文件是否已存在
	if _, err := os.Stat(configPath); err == nil && !force {
		fmt.Printf("配置文件 %s 已存在，跳过初始化\n", configPath)
		return nil
	}

	fmt.Printf("正在初始化配置文件: %s\n", configPath)

	// 创建默认配置
	defaultConfig := ci.createDefaultConfig()

	// 确保目录存在
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %v", err)
	}

	// 将配置写入文件
	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	fmt.Printf("✅ 配置文件初始化完成: %s\n", configPath)
	fmt.Println()
	fmt.Println("📝 请编辑配置文件，设置以下内容：")
	fmt.Println("   1. 腾讯云TTS: 在 tencent_cloud 部分填入您的 secret_id 和 secret_key")
	fmt.Println("   2. Edge TTS: 无需配置，可直接使用")
	fmt.Println("   3. 其他参数: 根据需要调整音色、语速等参数")
	fmt.Println()

	return nil
}

// createDefaultConfig 创建默认配置
func (ci *ConfigInitializer) createDefaultConfig() *model.Config {
	return &model.Config{
		InputFile: "input.txt",
		TencentCloud: model.TencentCloudConfig{
			SecretID:  "your_secret_id",
			SecretKey: "your_secret_key",
			Region:    "ap-beijing",
		},
		TTS: model.TTSConfig{
			VoiceType:       101008, // 智琪 - 女声
			Volume:          5,
			Speed:           1.0,
			PrimaryLanguage: 1,
			SampleRate:      16000,
			Codec:           "mp3",
		},
		EdgeTTS: model.EdgeTTSConfig{
			Voice:  "zh-CN-XiaoyiNeural",
			Rate:   "+0%",
			Volume: "+0%",
			Pitch:  "+0Hz",
		},
		Audio: model.AudioConfig{
			OutputDir:       "output",
			TempDir:         "temp",
			FinalOutput:     "merged_audio.mp3",
			SilenceDuration: 0.5,
		},
		Concurrent: model.ConcurrentConfig{
			MaxWorkers: 5,
			RateLimit:  20,
			BatchSize:  10,
		},
	}
}

// CreateSampleInputFile 创建示例输入文件
func (ci *ConfigInitializer) CreateSampleInputFile(inputPath string) error {
	return ci.CreateSampleInputFileWithForce(inputPath, false)
}

// CreateSampleInputFileWithForce 创建示例输入文件（支持强制覆盖）
func (ci *ConfigInitializer) CreateSampleInputFileWithForce(inputPath string, force bool) error {
	// 检查文件是否已存在
	if _, err := os.Stat(inputPath); err == nil && !force {
		fmt.Printf("示例输入文件 %s 已存在，跳过创建\n", inputPath)
		return nil
	}

	fmt.Printf("正在创建示例输入文件: %s\n", inputPath)

	sampleContent := `欢迎使用TTS语音合成应用！

这是一个功能强大的文本转语音工具。
支持腾讯云TTS和Microsoft Edge TTS两种引擎。
Edge TTS完全免费，无需API密钥。

特殊字符处理示例：
**代理（Agents）**能基于用户输入自主决策执行流程。
\*\*转义字符\*\*也能正确处理。
AI Agent可以automatically处理various任务。

符号测试：！@#$%^&*()
括号测试：（中文括号）和(English brackets)

请编辑此文件，添加您要转换的文本内容。
每行文本将被转换为一个音频片段，最后自动合并。

开始使用：
1. 免费版本：./github.com/difyz9/markdown2tts edge -i input.txt
2. 腾讯云版本：./github.com/difyz9/markdown2tts tts -i input.txt
`

	err := os.WriteFile(inputPath, []byte(sampleContent), 0644)
	if err != nil {
		return fmt.Errorf("创建示例输入文件失败: %v", err)
	}

	fmt.Printf("✅ 示例输入文件创建完成: %s\n", inputPath)
	return nil
}

// ShowQuickStart 显示快速开始指南
func (ci *ConfigInitializer) ShowQuickStart() {
	fmt.Println()
	fmt.Println("🚀 快速开始指南:")
	fmt.Println()
	fmt.Println("方式一：免费Edge TTS（推荐新手）")
	fmt.Println("   ./github.com/difyz9/markdown2tts edge -i input.txt")
	fmt.Println()
	fmt.Println("方式二：腾讯云TTS（需要API密钥）")
	fmt.Println("   1. 编辑 config.yaml，填入腾讯云密钥")
	fmt.Println("   2. ./github.com/difyz9/markdown2tts tts -i input.txt")
	fmt.Println()
	fmt.Println("方式三：测试文本处理效果")
	fmt.Println("   go run test_text_processor.go input.txt")
	fmt.Println()
	fmt.Println("📖 更多信息请查看：")
	fmt.Println("   - README.md - 完整使用说明")
	fmt.Println("   - docs/special-chars-handling.md - 特殊字符处理说明")
	fmt.Println("   - docs/quick-start.md - 详细快速开始指南")
	fmt.Println()
}

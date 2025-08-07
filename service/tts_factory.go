package service

import (
	"fmt"
	"github.com/difyz9/markdown2tts/model"
)

// TTSProviderFactory TTS提供商工厂
type TTSProviderFactory struct{}

// CreateProvider 根据配置创建相应的TTS提供商
func (factory *TTSProviderFactory) CreateProvider(providerType string, config *model.Config) (TTSProvider, error) {
	switch providerType {
	case "tencent", "tencentcloud":
		return NewTencentTTSProvider(
			config.TencentCloud.SecretID,
			config.TencentCloud.SecretKey,
			config.TencentCloud.Region,
			config,
		)
	case "edge", "edgetts":
		return NewEdgeTTSProvider(config), nil
	default:
		return nil, fmt.Errorf("不支持的TTS提供商: %s", providerType)
	}
}

// CreateUnifiedService 创建统一的TTS服务
func CreateUnifiedTTSService(providerType string, config *model.Config) (*UnifiedTTSService, error) {
	factory := &TTSProviderFactory{}
	
	provider, err := factory.CreateProvider(providerType, config)
	if err != nil {
		return nil, err
	}
	
	return NewUnifiedTTSService(provider, config), nil
}

// Example: 使用统一接口的示例
func ExampleUsageUnifiedTTS(config *model.Config) error {
	// 创建腾讯云TTS服务
	tencentService, err := CreateUnifiedTTSService("tencent", config)
	if err != nil {
		return fmt.Errorf("创建腾讯云TTS服务失败: %v", err)
	}

	// 处理Markdown文件
	err = tencentService.ProcessMarkdownFile("input.md", "output/")
	if err != nil {
		return fmt.Errorf("腾讯云TTS处理Markdown文件失败: %v", err)
	}

	// 创建Edge TTS服务
	edgeService, err := CreateUnifiedTTSService("edge", config)
	if err != nil {
		return fmt.Errorf("创建Edge TTS服务失败: %v", err)
	}

	// 处理普通文本文件
	err = edgeService.ProcessInputFile("input.txt", "output/")
	if err != nil {
		return fmt.Errorf("Edge TTS处理文本文件失败: %v", err)
	}

	fmt.Println("所有语音合成任务完成！")
	return nil
}

package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/difyz9/edge-tts-go/pkg/communicate"
	"github.com/difyz9/markdown2tts/model"
)

// EdgeTTSProvider Edge TTS提供商
type EdgeTTSProvider struct {
	config *model.Config
}

// NewEdgeTTSProvider 创建Edge TTS提供商
func NewEdgeTTSProvider(config *model.Config) *EdgeTTSProvider {
	return &EdgeTTSProvider{
		config: config,
	}
}

// GenerateAudio 生成音频
func (etp *EdgeTTSProvider) GenerateAudio(ctx context.Context, text string, index int) (string, error) {
	// 使用配置中的语音参数
	voice := etp.config.EdgeTTS.Voice
	if voice == "" {
		voice = "zh-CN-XiaoyiNeural" // 默认中文女声
	}

	rate := etp.config.EdgeTTS.Rate
	if rate == "" {
		rate = "+0%" // 默认正常语速
	}

	volume := etp.config.EdgeTTS.Volume
	if volume == "" {
		volume = "+0%" // 默认正常音量
	}

	pitch := etp.config.EdgeTTS.Pitch
	if pitch == "" {
		pitch = "+0Hz" // 默认正常音调
	}

	// 创建Edge TTS通信实例
	comm, err := communicate.NewCommunicate(
		text,
		voice,
		rate,   // rate - 语速
		volume, // volume - 音量
		pitch,  // pitch - 音调
		"",     // proxy
		10,     // connectTimeout
		60,     // receiveTimeout
	)
	if err != nil {
		return "", fmt.Errorf("创建Edge TTS通信失败: %v", err)
	}

	// 生成文件名
	filename := fmt.Sprintf("audio_%03d.mp3", index)
	audioPath := filepath.Join(etp.config.Audio.TempDir, filename)

	// 保存音频文件
	err = comm.Save(ctx, audioPath, "")
	if err != nil {
		return "", fmt.Errorf("保存音频文件失败: %v", err)
	}

	// 验证生成的音频文件
	if err := etp.validateAudioFile(audioPath); err != nil {
		// 删除无效的音频文件
		os.Remove(audioPath)
		return "", fmt.Errorf("音频文件验证失败: %v", err)
	}

	return audioPath, nil
}

// GetProviderName 获取提供商名称
func (etp *EdgeTTSProvider) GetProviderName() string {
	return "EdgeTTS"
}

// ValidateConfig 验证配置是否正确
func (etp *EdgeTTSProvider) ValidateConfig() error {
	// Edge TTS 不需要特殊配置验证，使用默认值即可
	return nil
}

// GetMaxTextLength 获取单次请求最大文本长度
func (etp *EdgeTTSProvider) GetMaxTextLength() int {
	return 1000 // Edge TTS 支持较长文本，设置为1000字符
}

// GetRecommendedRateLimit 获取推荐的速率限制（每秒请求数）
func (etp *EdgeTTSProvider) GetRecommendedRateLimit() int {
	return 10 // Edge TTS 可以支持更高的并发，设置为每秒10个请求
}

// validateAudioFile 验证音频文件的有效性
func (etp *EdgeTTSProvider) validateAudioFile(audioPath string) error {
	// 检查文件是否存在
	fileInfo, err := os.Stat(audioPath)
	if err != nil {
		return fmt.Errorf("音频文件不存在: %v", err)
	}

	// 检查文件大小（MP3文件通常至少几KB）
	const minFileSize = 1024 // 最小1KB
	if fileInfo.Size() < minFileSize {
		return fmt.Errorf("音频文件过小 (%d bytes)，可能为空或损坏", fileInfo.Size())
	}

	// 检查文件是否可读
	file, err := os.Open(audioPath)
	if err != nil {
		return fmt.Errorf("无法打开音频文件: %v", err)
	}
	defer file.Close()

	// 读取文件头部，检查是否为有效的MP3文件
	buffer := make([]byte, 10)
	n, err := file.Read(buffer)
	if err != nil || n < 3 {
		return fmt.Errorf("无法读取音频文件头部")
	}

	// 检查MP3文件头部标识
	// MP3文件通常以ID3标签 (ID3) 或 MP3帧同步字 (0xFF 0xFB/0xFA/0xF3/0xF2) 开头
	if n >= 3 && (string(buffer[:3]) == "ID3" ||
		(buffer[0] == 0xFF && (buffer[1]&0xF0) == 0xF0)) {
		fmt.Printf("  ✓ 音频文件验证通过: %s (%.2f KB)\n", audioPath, float64(fileInfo.Size())/1024)
		return nil
	}

	return fmt.Errorf("音频文件格式无效，可能不是有效的MP3文件")
}

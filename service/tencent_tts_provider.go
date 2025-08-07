package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/difyz9/markdown2tts/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tts/v20190823"
)

// TencentTTSProvider 腾讯云TTS提供商
type TencentTTSProvider struct {
	client *tts.Client
	config *model.Config
}

// NewTencentTTSProvider 创建腾讯云TTS提供商
func NewTencentTTSProvider(secretId, secretKey, region string, config *model.Config) (*TencentTTSProvider, error) {
	// 实例化一个认证对象
	credential := common.NewCredential(secretId, secretKey)
	
	// 实例化一个客户端配置对象
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tts.tencentcloudapi.com"

	// 实例化要请求产品的client对象
	client, err := tts.NewClient(credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("创建腾讯云TTS客户端失败: %v", err)
	}

	provider := &TencentTTSProvider{
		client: client,
		config: config,
	}

	// 验证配置
	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// GenerateAudio 生成音频
func (ttp *TencentTTSProvider) GenerateAudio(ctx context.Context, text string, index int) (string, error) {
	// 创建TTS请求
	req := &model.TTSRequest{
		Text:            text,
		VoiceType:       ttp.config.TTS.VoiceType,
		Volume:          ttp.config.TTS.Volume,
		Speed:           ttp.config.TTS.Speed,
		PrimaryLanguage: ttp.config.TTS.PrimaryLanguage,
		SampleRate:      ttp.config.TTS.SampleRate,
		Codec:           ttp.config.TTS.Codec,
	}

	// 创建TTS任务
	response, err := ttp.createTTSTask(req)
	if err != nil {
		return "", fmt.Errorf("创建TTS任务失败: %v", err)
	}

	if !response.Success {
		return "", fmt.Errorf("TTS任务创建失败: %s", response.Error)
	}

	// 等待任务完成并下载音频
	audioPath, err := ttp.waitForTaskAndDownload(ctx, response.TaskID, index)
	if err != nil {
		return "", fmt.Errorf("下载音频失败: %v", err)
	}

	return audioPath, nil
}

// GetProviderName 获取提供商名称
func (ttp *TencentTTSProvider) GetProviderName() string {
	return "TencentCloud"
}

// ValidateConfig 验证配置是否正确
func (ttp *TencentTTSProvider) ValidateConfig() error {
	if ttp.config.TencentCloud.SecretID == "" || ttp.config.TencentCloud.SecretID == "your_secret_id" {
		return fmt.Errorf("腾讯云SecretID未配置")
	}
	if ttp.config.TencentCloud.SecretKey == "" || ttp.config.TencentCloud.SecretKey == "your_secret_key" {
		return fmt.Errorf("腾讯云SecretKey未配置")
	}
	if ttp.config.TencentCloud.Region == "" {
		return fmt.Errorf("腾讯云Region未配置")
	}
	return nil
}

// GetMaxTextLength 获取单次请求最大文本长度
func (ttp *TencentTTSProvider) GetMaxTextLength() int {
	return 150 // 腾讯云TTS单次最大150个字符
}

// GetRecommendedRateLimit 获取推荐的速率限制（每秒请求数）
func (ttp *TencentTTSProvider) GetRecommendedRateLimit() int {
	return 5 // 腾讯云TTS建议每秒不超过5个请求
}

// createTTSTask 创建TTS任务
func (ttp *TencentTTSProvider) createTTSTask(req *model.TTSRequest) (*model.TTSResponse, error) {
	// 设置默认值
	if req.VoiceType == 0 {
		req.VoiceType = 101008 // 智琪 - 女声
	}
	if req.Volume == 0 {
		req.Volume = 5
	}
	if req.Speed == 0 {
		req.Speed = 1.0 // 腾讯云TTS速度范围：0.6-1.5，默认1.0
	}
	if req.PrimaryLanguage == 0 {
		req.PrimaryLanguage = 1
	}
	if req.SampleRate == 0 {
		req.SampleRate = 16000
	}
	if req.Codec == "" {
		req.Codec = "mp3"
	}

	// 实例化一个请求对象
	request := tts.NewCreateTtsTaskRequest()
	request.Text = common.StringPtr(req.Text)
	request.Volume = common.Float64Ptr(float64(req.Volume))
	request.Speed = common.Float64Ptr(req.Speed)
	request.VoiceType = common.Int64Ptr(req.VoiceType)
	request.PrimaryLanguage = common.Int64Ptr(req.PrimaryLanguage)
	request.SampleRate = common.Uint64Ptr(uint64(req.SampleRate))
	request.Codec = common.StringPtr(req.Codec)

	// 发起请求
	response, err := ttp.client.CreateTtsTask(request)
	if err != nil {
		return &model.TTSResponse{
			Success: false,
			Error:   fmt.Sprintf("调用腾讯云TTS失败: %v", err),
		}, nil
	}

	return &model.TTSResponse{
		Success: true,
		TaskID:  *response.Response.Data.TaskId,
		Message: "TTS任务创建成功",
	}, nil
}

// describeTTSTaskStatus 查询TTS任务状态
func (ttp *TencentTTSProvider) describeTTSTaskStatus(taskID string) (*model.TTSStatusResponse, error) {
	// 实例化一个请求对象
	request := tts.NewDescribeTtsTaskStatusRequest()
	request.TaskId = common.StringPtr(taskID)

	// 发起请求
	response, err := ttp.client.DescribeTtsTaskStatus(request)
	if err != nil {
		return &model.TTSStatusResponse{
			Success: false,
			Error:   fmt.Sprintf("查询TTS任务状态失败: %v", err),
		}, nil
	}

	result := &model.TTSStatusResponse{
		Success:   true,
		Status:    *response.Response.Data.Status,
		StatusStr: *response.Response.Data.StatusStr,
	}

	if response.Response.Data.ResultUrl != nil {
		result.AudioURL = *response.Response.Data.ResultUrl
	}

	if response.Response.Data.ErrorMsg != nil {
		result.ErrorMsg = *response.Response.Data.ErrorMsg
	}

	return result, nil
}

// waitForTaskAndDownload 等待任务完成并下载音频
func (ttp *TencentTTSProvider) waitForTaskAndDownload(ctx context.Context, taskID string, index int) (string, error) {
	// 轮询任务状态
	maxWaitTime := 60 * time.Second // 最大等待60秒
	checkInterval := 2 * time.Second // 每2秒检查一次
	startTime := time.Now()

	for time.Since(startTime) < maxWaitTime {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		status, err := ttp.describeTTSTaskStatus(taskID)
		if err != nil {
			return "", fmt.Errorf("查询任务状态失败: %v", err)
		}

		if !status.Success {
			return "", fmt.Errorf("查询任务状态失败: %s", status.Error)
		}

		switch status.Status {
		case 2: // 任务完成
			if status.AudioURL == "" {
				return "", fmt.Errorf("任务完成但没有获取到音频URL")
			}
			// 下载音频文件
			return ttp.downloadAudio(status.AudioURL, index)

		case 3: // 任务失败
			return "", fmt.Errorf("TTS任务失败: %s", status.ErrorMsg)

		case 0, 1: // 任务排队中或处理中
			fmt.Printf("  ⏳ 任务 %d 状态: %s, 等待中...\n", index, status.StatusStr)
			time.Sleep(checkInterval)

		default:
			return "", fmt.Errorf("未知任务状态: %d", status.Status)
		}
	}

	return "", fmt.Errorf("任务超时，等待时间超过 %v", maxWaitTime)
}

// downloadAudio 下载音频文件
func (ttp *TencentTTSProvider) downloadAudio(audioURL string, index int) (string, error) {
	// 生成文件名
	filename := fmt.Sprintf("audio_%03d.mp3", index)
	audioPath := filepath.Join(ttp.config.Audio.TempDir, filename)

	// 下载文件
	resp, err := http.Get(audioURL)
	if err != nil {
		return "", fmt.Errorf("下载音频失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载音频失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 创建本地文件
	file, err := os.Create(audioPath)
	if err != nil {
		return "", fmt.Errorf("创建音频文件失败: %v", err)
	}
	defer file.Close()

	// 复制数据
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("保存音频文件失败: %v", err)
	}

	// 验证生成的音频文件
	if err := ttp.validateAudioFile(audioPath); err != nil {
		// 删除无效的音频文件
		os.Remove(audioPath)
		return "", fmt.Errorf("音频文件验证失败: %v", err)
	}

	return audioPath, nil
}

// validateAudioFile 验证音频文件的有效性
func (ttp *TencentTTSProvider) validateAudioFile(audioPath string) error {
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

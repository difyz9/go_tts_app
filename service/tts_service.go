package service

import (
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tts/v20190823"
	"os"
	"github.com/difyz9/markdown2tts/model"
)

type TTSService struct {
	client *tts.Client
}

func NewTTSService(secretId, secretKey, region string) *TTSService {

	// 实例化一个认证对象
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	// 实例化一个客户端配置对象
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "tts.tencentcloudapi.com"

	// 实例化要请求产品的client对象
	client, err := tts.NewClient(credential, region, cpf)
	if err != nil {
		fmt.Println("创建腾讯云TTS客户端失败:", err)
		return nil
	}

	return &TTSService{
		client: client,
	}
}

// 创建TTS任务
func (s *TTSService) CreateTTSTask(req *model.TTSRequest) (*model.TTSResponse, error) {
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
	response, err := s.client.CreateTtsTask(request)
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

// 查询TTS任务状态
func (s *TTSService) DescribeTTSTaskStatus(taskID string) (*model.TTSStatusResponse, error) {
	// 实例化一个请求对象
	request := tts.NewDescribeTtsTaskStatusRequest()
	request.TaskId = common.StringPtr(taskID)

	// 发起请求
	response, err := s.client.DescribeTtsTaskStatus(request)
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

// EnsureDir 确保目录存在，如果不存在则创建
func EnsureDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

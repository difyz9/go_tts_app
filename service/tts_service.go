package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/difyz9/markdown2tts/model"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tts/v20190823"
	"golang.org/x/time/rate"
)

type TTSService struct {
	client        *tts.Client
	config        *model.Config
	limiter       *rate.Limiter
	textProcessor *TextProcessor
}

func NewTTSService(secretId, secretKey, region string, config *model.Config) *TTSService {
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

	// 创建速率限制器，腾讯云TTS有配额限制，设置较保守的限制
	rateLimit := rate.Every(time.Second / time.Duration(config.Concurrent.RateLimit))
	limiter := rate.NewLimiter(rateLimit, config.Concurrent.RateLimit)

	return &TTSService{
		client:        client,
		config:        config,
		limiter:       limiter,
		textProcessor: NewTextProcessor(),
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

// ProcessMarkdownFile 使用智能Markdown解析处理文件
func (s *TTSService) ProcessMarkdownFile(inputFile, outputDir string) error {
	// 确保目录存在
	if err := os.MkdirAll(s.config.Audio.TempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 读取文件内容
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 使用专业Markdown处理器提取文本
	sentences := s.textProcessor.ProcessMarkdownDocument(string(content))

	if len(sentences) == 0 {
		return fmt.Errorf("没有提取到有效的文本内容")
	}

	fmt.Printf("📊 Markdown处理统计: 提取到 %d 个有效句子\n", len(sentences))

	// 创建任务
	var tasks []TTSTask
	for i, sentence := range sentences {
		tasks = append(tasks, TTSTask{Index: i, Text: sentence})
	}

	// 并发处理任务
	results, err := s.processTTSTasksConcurrent(tasks)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return fmt.Errorf("没有成功生成任何音频文件")
	}

	// 按索引排序结果，确保音频文件按原始顺序合并
	sort.Slice(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})

	// 收集所有音频文件
	audioFiles := make([]string, 0, len(results))
	for _, result := range results {
		audioFiles = append(audioFiles, result.AudioFile)
	}

	// 合并音频文件
	return s.mergeAudioFiles(audioFiles)
}

// ProcessInputFileConcurrent 并发处理输入文件
func (s *TTSService) ProcessInputFileConcurrent() error {
	// 确保目录存在
	if err := os.MkdirAll(s.config.Audio.TempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	if err := os.MkdirAll(s.config.Audio.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 读取输入文件
	lines, err := s.readInputFile()
	if err != nil {
		return err
	}

	fmt.Printf("读取到 %d 行文本，开始并发生成音频...\n", len(lines))
	fmt.Printf("并发配置: workers=%d, rate_limit=%d/秒, batch_size=%d\n",
		s.config.Concurrent.MaxWorkers,
		s.config.Concurrent.RateLimit,
		s.config.Concurrent.BatchSize)

	// 创建任务列表
	tasks := make([]TTSTask, 0, len(lines))
	emptyLineCount := 0
	invalidTextCount := 0

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 跳过完全空行
		if trimmedLine == "" {
			emptyLineCount++
			continue
		}

		// 跳过只包含空白字符的行
		if len(strings.ReplaceAll(strings.ReplaceAll(trimmedLine, " ", ""), "\t", "")) == 0 {
			emptyLineCount++
			continue
		}

		// 使用文本处理器验证文本
		if !s.textProcessor.IsValidTextForTTS(trimmedLine) {
			invalidTextCount++
			continue
		}

		tasks = append(tasks, TTSTask{Index: i, Text: line})
	}

	if len(tasks) == 0 {
		return fmt.Errorf("没有有效的文本行需要处理")
	}

	fmt.Printf("📊 文本处理统计: 总行数=%d, 空行=%d, 无效文本=%d, 有效任务=%d\n",
		len(lines), emptyLineCount, invalidTextCount, len(tasks))

	// 并发处理任务
	results, err := s.processTTSTasksConcurrent(tasks)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return fmt.Errorf("没有成功生成任何音频文件")
	}

	// 按索引排序结果，确保音频文件按原始顺序合并
	sort.Slice(results, func(i, j int) bool {
		return results[i].Index < results[j].Index
	})

	// 收集所有音频文件
	audioFiles := make([]string, 0, len(results))
	for _, result := range results {
		audioFiles = append(audioFiles, result.AudioFile)
	}

	// 合并音频文件
	return s.mergeAudioFiles(audioFiles)
}

// readInputFile 读取输入文件
func (s *TTSService) readInputFile() ([]string, error) {
	file, err := os.Open(s.config.InputFile)
	if err != nil {
		return nil, fmt.Errorf("打开输入文件失败: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取输入文件失败: %v", err)
	}

	return lines, nil
}

// processTTSTasksConcurrent 并发处理TTS任务
func (s *TTSService) processTTSTasksConcurrent(tasks []TTSTask) ([]TTSResult, error) {
	// 创建通道
	taskChan := make(chan TTSTask, len(tasks))
	resultChan := make(chan TTSResult, len(tasks))

	// 将任务发送到通道
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 确定worker数量
	workerCount := s.config.Concurrent.MaxWorkers
	if workerCount > len(tasks) {
		workerCount = len(tasks)
	}

	fmt.Printf("启动 %d 个worker开始处理...\n", workerCount)

	// 启动workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go s.ttsWorker(i, taskChan, resultChan, &wg)
	}

	// 等待所有workers完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var results []TTSResult
	successCount := 0
	failureCount := 0

	for result := range resultChan {
		results = append(results, result)
		if result.Error != nil {
			failureCount++
			fmt.Printf("✗ 任务 %d 失败: %v\n", result.Index, result.Error)
		} else {
			successCount++
			fmt.Printf("✓ 任务 %d 完成: %s\n", result.Index, result.AudioFile)
		}
	}

	fmt.Printf("\n处理完成: 成功 %d, 失败 %d\n\n", successCount, failureCount)

	return results, nil
}

// ttsWorker 腾讯云TTS工作协程
func (s *TTSService) ttsWorker(workerID int, taskChan <-chan TTSTask, resultChan chan<- TTSResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskChan {
		fmt.Printf("Worker %d 处理任务 %d: %s\n", workerID, task.Index, task.Text)

		// 限制请求频率
		err := s.limiter.Wait(context.Background())
		if err != nil {
			resultChan <- TTSResult{
				Index: task.Index,
				Error: fmt.Errorf("等待速率限制失败: %v", err),
			}
			continue
		}

		// 生成音频，带重试机制
		audioFile, err := s.generateAudioWithRetry(task.Text, task.Index, 3)
		resultChan <- TTSResult{
			Index:     task.Index,
			AudioFile: audioFile,
			Error:     err,
		}
	}
}

// generateAudioWithRetry 带重试机制的音频生成
func (s *TTSService) generateAudioWithRetry(text string, index int, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		audioPath, err := s.generateAudioForText(text, index)
		if err == nil {
			if attempt > 1 {
				fmt.Printf("  ✓ 任务 %d 重试第 %d 次成功\n", index, attempt-1)
			}
			return audioPath, nil
		}

		lastErr = err
		fmt.Printf("  ✗ 任务 %d 第 %d 次尝试失败: %v\n", index, attempt, err)

		if attempt < maxRetries {
			// 等待后重试，递增等待时间
			waitTime := time.Duration(attempt) * time.Second * 2 // 腾讯云需要更长等待时间
			fmt.Printf("  ⏳ 任务 %d 等待 %v 后重试...\n", index, waitTime)
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("任务 %d 经过 %d 次重试后仍然失败，最后错误: %v", index, maxRetries, lastErr)
}

// generateAudioForText 为文本生成音频
func (s *TTSService) generateAudioForText(text string, index int) (string, error) {
	// 处理文本：去除特殊字符和格式
	processedText := s.textProcessor.ProcessText(text)
	if strings.TrimSpace(processedText) == "" {
		return "", fmt.Errorf("处理后的文本为空")
	}

	// 如果处理前后不同，显示处理效果
	if processedText != text {
		fmt.Printf("  📝 文本处理: \"%s\" → \"%s\"\n", text, processedText)
	}

	// 创建TTS请求
	req := &model.TTSRequest{
		Text:            processedText,
		VoiceType:       s.config.TTS.VoiceType,
		Volume:          s.config.TTS.Volume,
		Speed:           s.config.TTS.Speed,
		PrimaryLanguage: s.config.TTS.PrimaryLanguage,
		SampleRate:      s.config.TTS.SampleRate,
		Codec:           s.config.TTS.Codec,
	}

	// 创建TTS任务
	response, err := s.CreateTTSTask(req)
	if err != nil {
		return "", fmt.Errorf("创建TTS任务失败: %v", err)
	}

	if !response.Success {
		return "", fmt.Errorf("TTS任务创建失败: %s", response.Error)
	}

	// 等待任务完成并下载音频
	audioPath, err := s.waitForTaskAndDownload(response.TaskID, index)
	if err != nil {
		return "", fmt.Errorf("下载音频失败: %v", err)
	}

	return audioPath, nil
}

// waitForTaskAndDownload 等待任务完成并下载音频
func (s *TTSService) waitForTaskAndDownload(taskID string, index int) (string, error) {
	// 轮询任务状态
	maxWaitTime := 60 * time.Second // 最大等待60秒
	checkInterval := 2 * time.Second // 每2秒检查一次
	startTime := time.Now()

	for time.Since(startTime) < maxWaitTime {
		status, err := s.DescribeTTSTaskStatus(taskID)
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
			return s.downloadAudio(status.AudioURL, index)

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
func (s *TTSService) downloadAudio(audioURL string, index int) (string, error) {
	// 生成文件名
	filename := fmt.Sprintf("audio_%03d.mp3", index)
	audioPath := filepath.Join(s.config.Audio.TempDir, filename)

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
	if err := s.validateAudioFile(audioPath); err != nil {
		// 删除无效的音频文件
		os.Remove(audioPath)
		return "", fmt.Errorf("音频文件验证失败: %v", err)
	}

	return audioPath, nil
}

// validateAudioFile 验证音频文件的有效性
func (s *TTSService) validateAudioFile(audioPath string) error {
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

// mergeAudioFiles 合并音频文件
func (s *TTSService) mergeAudioFiles(audioFiles []string) error {
	if len(audioFiles) == 0 {
		return fmt.Errorf("没有音频文件需要合并")
	}

	fmt.Printf("开始合并 %d 个音频文件...\n", len(audioFiles))

	// 预先验证所有音频文件
	validAudioFiles := []string{}
	invalidCount := 0

	for _, audioFile := range audioFiles {
		if err := s.validateAudioFile(audioFile); err != nil {
			fmt.Printf("⚠️  跳过无效音频文件: %s, 原因: %v\n", audioFile, err)
			invalidCount++
			// 删除无效文件
			os.Remove(audioFile)
			continue
		}
		validAudioFiles = append(validAudioFiles, audioFile)
	}

	if len(validAudioFiles) == 0 {
		return fmt.Errorf("没有有效的音频文件可以合并")
	}

	if invalidCount > 0 {
		fmt.Printf("📊 音频文件验证统计: 有效 %d, 无效 %d\n", len(validAudioFiles), invalidCount)
	}

	// 输出文件路径
	outputPath := filepath.Join(s.config.Audio.OutputDir, s.config.Audio.FinalOutput)

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outputFile.Close()

	// 逐个读取并合并音频文件
	for i, audioFile := range validAudioFiles {
		fmt.Printf("合并文件 %d/%d: %s\n", i+1, len(validAudioFiles), audioFile)

		inputFile, err := os.Open(audioFile)
		if err != nil {
			return fmt.Errorf("打开音频文件失败 %s: %v", audioFile, err)
		}

		// 复制文件内容
		_, err = outputFile.ReadFrom(inputFile)
		inputFile.Close()

		if err != nil {
			return fmt.Errorf("复制音频文件失败 %s: %v", audioFile, err)
		}
	}

	fmt.Printf("音频合并完成: %s\n", outputPath)
	return nil
}

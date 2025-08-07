package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/difyz9/markdown2tts/model"
	"golang.org/x/time/rate"
)

// TTSProvider 语音合成提供商接口
type TTSProvider interface {
	// GenerateAudio 生成音频，返回音频文件路径
	GenerateAudio(ctx context.Context, text string, index int) (string, error)
	
	// GetProviderName 获取提供商名称
	GetProviderName() string
	
	// ValidateConfig 验证配置是否正确
	ValidateConfig() error
	
	// GetMaxTextLength 获取单次请求最大文本长度
	GetMaxTextLength() int
	
	// GetRecommendedRateLimit 获取推荐的速率限制（每秒请求数）
	GetRecommendedRateLimit() int
}

// UnifiedTTSTask 统一的TTS任务结构
type UnifiedTTSTask struct {
	Index int
	Text  string
}

// UnifiedTTSResult 统一的TTS任务结果
type UnifiedTTSResult struct {
	Index     int
	AudioFile string
	Error     error
}

// UnifiedTTSService 统一的TTS服务
type UnifiedTTSService struct {
	provider      TTSProvider
	config        *model.Config
	limiter       *rate.Limiter
	textProcessor *TextProcessor
}

// NewUnifiedTTSService 创建统一的TTS服务
func NewUnifiedTTSService(provider TTSProvider, config *model.Config) *UnifiedTTSService {
	// 使用提供商推荐的速率限制，如果配置中没有设置的话
	rateLimit := config.Concurrent.RateLimit
	if rateLimit <= 0 {
		rateLimit = provider.GetRecommendedRateLimit()
	}
	
	// 创建速率限制器
	rateLimiter := rate.Every(time.Second / time.Duration(rateLimit))
	limiter := rate.NewLimiter(rateLimiter, rateLimit)

	return &UnifiedTTSService{
		provider:      provider,
		config:        config,
		limiter:       limiter,
		textProcessor: NewTextProcessor(),
	}
}

// ProcessText 统一的文本处理
func (uts *UnifiedTTSService) ProcessText(text string) (string, error) {
	// 使用文本处理器处理文本
	processedText := uts.textProcessor.ProcessText(text)
	
	// 检查文本长度是否超过提供商限制
	maxLength := uts.provider.GetMaxTextLength()
	if maxLength > 0 && len(processedText) > maxLength {
		// 如果超过长度限制，进行智能分割
		return uts.textProcessor.SplitTextIntelligently(processedText, maxLength), nil
	}
	
	return processedText, nil
}

// GenerateAudioWithRateLimit 带速率限制的音频生成
func (uts *UnifiedTTSService) GenerateAudioWithRateLimit(ctx context.Context, text string, index int) (string, error) {
	// 等待速率限制
	if err := uts.limiter.Wait(ctx); err != nil {
		return "", err
	}
	
	// 处理文本
	processedText, err := uts.ProcessText(text)
	if err != nil {
		return "", err
	}
	
	// 调用提供商生成音频
	return uts.provider.GenerateAudio(ctx, processedText, index)
}

// ProcessMarkdownFile 处理Markdown文件
func (uts *UnifiedTTSService) ProcessMarkdownFile(inputFile, outputDir string) error {
	return uts.processFile(inputFile, outputDir, true)
}

// ProcessInputFile 处理普通文本文件
func (uts *UnifiedTTSService) ProcessInputFile(inputFile, outputDir string) error {
	return uts.processFile(inputFile, outputDir, false)
}

// processFile 通用文件处理方法
func (uts *UnifiedTTSService) processFile(inputFile, outputDir string, isMarkdown bool) error {
	// 确保目录存在
	if err := os.MkdirAll(uts.config.Audio.TempDir, 0755); err != nil {
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

	var sentences []string
	if isMarkdown {
		// 使用专业Markdown处理器提取文本
		sentences = uts.textProcessor.ProcessMarkdownDocument(string(content))
	} else {
		// 逐行处理普通文本文件
		lines, err := uts.readInputFile(inputFile)
		if err != nil {
			return err
		}
		sentences = uts.filterValidLines(lines)
	}

	if len(sentences) == 0 {
		return fmt.Errorf("没有提取到有效的文本内容")
	}

	fmt.Printf("📊 文本处理统计 [%s]: 提取到 %d 个有效句子\n", uts.provider.GetProviderName(), len(sentences))

	// 创建任务
	var tasks []UnifiedTTSTask
	for i, sentence := range sentences {
		tasks = append(tasks, UnifiedTTSTask{Index: i, Text: sentence})
	}

	// 并发处理任务
	results, err := uts.processTTSTasksConcurrent(tasks)
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
	return uts.mergeAudioFiles(audioFiles)
}

// readInputFile 读取输入文件
func (uts *UnifiedTTSService) readInputFile(inputFile string) ([]string, error) {
	file, err := os.Open(inputFile)
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

// filterValidLines 过滤有效的文本行
func (uts *UnifiedTTSService) filterValidLines(lines []string) []string {
	var validLines []string
	emptyLineCount := 0
	invalidTextCount := 0

	for _, line := range lines {
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
		if !uts.textProcessor.IsValidTextForTTS(trimmedLine) {
			invalidTextCount++
			continue
		}

		validLines = append(validLines, line)
	}

	fmt.Printf("📊 文本过滤统计: 总行数=%d, 空行=%d, 无效文本=%d, 有效行数=%d\n",
		len(lines), emptyLineCount, invalidTextCount, len(validLines))

	return validLines
}

// processTTSTasksConcurrent 并发处理TTS任务
func (uts *UnifiedTTSService) processTTSTasksConcurrent(tasks []UnifiedTTSTask) ([]UnifiedTTSResult, error) {
	// 创建通道
	taskChan := make(chan UnifiedTTSTask, len(tasks))
	resultChan := make(chan UnifiedTTSResult, len(tasks))

	// 将任务发送到通道
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 确定worker数量
	workerCount := uts.config.Concurrent.MaxWorkers
	if workerCount > len(tasks) {
		workerCount = len(tasks)
	}

	fmt.Printf("启动 %d 个worker开始处理 [%s]...\n", workerCount, uts.provider.GetProviderName())
	fmt.Printf("并发配置: workers=%d, rate_limit=%d/秒, provider=%s\n",
		workerCount,
		uts.config.Concurrent.RateLimit,
		uts.provider.GetProviderName())

	// 启动workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go uts.ttsWorker(i, taskChan, resultChan, &wg)
	}

	// 等待所有workers完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var results []UnifiedTTSResult
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

	fmt.Printf("\n处理完成 [%s]: 成功 %d, 失败 %d\n\n", uts.provider.GetProviderName(), successCount, failureCount)

	return results, nil
}

// ttsWorker TTS工作协程
func (uts *UnifiedTTSService) ttsWorker(workerID int, taskChan <-chan UnifiedTTSTask, resultChan chan<- UnifiedTTSResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskChan {
		fmt.Printf("Worker %d 处理任务 %d [%s]: %s\n", workerID, task.Index, uts.provider.GetProviderName(), task.Text)

		// 生成音频，带重试机制
		audioFile, err := uts.generateAudioWithRetry(task.Text, task.Index, 3)
		resultChan <- UnifiedTTSResult{
			Index:     task.Index,
			AudioFile: audioFile,
			Error:     err,
		}
	}
}

// generateAudioWithRetry 带重试机制的音频生成
func (uts *UnifiedTTSService) generateAudioWithRetry(text string, index int, maxRetries int) (string, error) {
	var lastErr error
	ctx := context.Background()

	for attempt := 1; attempt <= maxRetries; attempt++ {
		audioPath, err := uts.GenerateAudioWithRateLimit(ctx, text, index)
		if err == nil {
			if attempt > 1 {
				fmt.Printf("  ✓ 任务 %d 重试第 %d 次成功 [%s]\n", index, attempt-1, uts.provider.GetProviderName())
			}
			return audioPath, nil
		}

		lastErr = err
		fmt.Printf("  ✗ 任务 %d 第 %d 次尝试失败 [%s]: %v\n", index, attempt, uts.provider.GetProviderName(), err)

		if attempt < maxRetries {
			// 等待后重试，递增等待时间
			waitTime := time.Duration(attempt) * time.Second
			fmt.Printf("  ⏳ 任务 %d 等待 %v 后重试 [%s]...\n", index, waitTime, uts.provider.GetProviderName())
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("任务 %d 经过 %d 次重试后仍然失败 [%s]，最后错误: %v", index, maxRetries, uts.provider.GetProviderName(), lastErr)
}

// mergeAudioFiles 合并音频文件
func (uts *UnifiedTTSService) mergeAudioFiles(audioFiles []string) error {
	if len(audioFiles) == 0 {
		return fmt.Errorf("没有音频文件需要合并")
	}

	fmt.Printf("开始合并 %d 个音频文件 [%s]...\n", len(audioFiles), uts.provider.GetProviderName())

	// 预先验证所有音频文件
	validAudioFiles := []string{}
	invalidCount := 0

	for _, audioFile := range audioFiles {
		if err := uts.validateAudioFile(audioFile); err != nil {
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
		fmt.Printf("📊 音频文件验证统计 [%s]: 有效 %d, 无效 %d\n", uts.provider.GetProviderName(), len(validAudioFiles), invalidCount)
	}

	// 输出文件路径
	outputPath := filepath.Join(uts.config.Audio.OutputDir, uts.config.Audio.FinalOutput)

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outputFile.Close()

	// 逐个读取并合并音频文件
	for i, audioFile := range validAudioFiles {
		fmt.Printf("合并文件 %d/%d [%s]: %s\n", i+1, len(validAudioFiles), uts.provider.GetProviderName(), audioFile)

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

	fmt.Printf("音频合并完成 [%s]: %s\n", uts.provider.GetProviderName(), outputPath)
	return nil
}

// validateAudioFile 验证音频文件的有效性
func (uts *UnifiedTTSService) validateAudioFile(audioPath string) error {
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
		return nil
	}

	return fmt.Errorf("音频文件格式无效，可能不是有效的MP3文件")
}

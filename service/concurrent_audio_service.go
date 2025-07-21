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
	"tts_app/model"

	"golang.org/x/time/rate"
)

// TTSTask TTS任务结构
type TTSTask struct {
	Index int
	Text  string
}

// TTSResult TTS任务结果
type TTSResult struct {
	Index     int
	AudioFile string
	Error     error
}

// ConcurrentAudioService 并发音频服务
type ConcurrentAudioService struct {
	config        *model.Config
	ttsService    *TTSService
	limiter       *rate.Limiter
	textProcessor *TextProcessor
}

// NewConcurrentAudioService 创建并发音频服务
func NewConcurrentAudioService(config *model.Config, ttsService *TTSService) *ConcurrentAudioService {
	// 创建速率限制器，限制为每秒不超过配置的请求数
	rateLimit := rate.Every(time.Second / time.Duration(config.Concurrent.RateLimit))
	limiter := rate.NewLimiter(rateLimit, config.Concurrent.RateLimit)

	return &ConcurrentAudioService{
		config:        config,
		ttsService:    ttsService,
		limiter:       limiter,
		textProcessor: NewTextProcessor(),
	}
}

// ProcessInputFileConcurrent 并发处理历史文件
func (cas *ConcurrentAudioService) ProcessInputFileConcurrent() error {
	// 确保目录存在
	if err := os.MkdirAll(cas.config.Audio.TempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	if err := os.MkdirAll(cas.config.Audio.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 读取历史文件
	lines, err := cas.readInputFile()
	if err != nil {
		return err
	}

	fmt.Printf("读取到 %d 行文本，开始并发生成音频...\n", len(lines))
	fmt.Printf("并发配置: workers=%d, rate_limit=%d/秒, batch_size=%d\n",
		cas.config.Concurrent.MaxWorkers,
		cas.config.Concurrent.RateLimit,
		cas.config.Concurrent.BatchSize)

	// 创建任务列表
	tasks := make([]TTSTask, 0, len(lines))
	validLineCount := 0
	skippedLineCount := 0
	
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			skippedLineCount++
			continue // 跳过空行
		}

		// 快速过滤明显的标记行（仅针对行首的标记）
		if strings.HasPrefix(trimmedLine, "## ") ||
			strings.HasPrefix(trimmedLine, "### ") ||
			strings.HasPrefix(trimmedLine, "#### ") ||
			strings.HasPrefix(trimmedLine, "** ") ||
			strings.HasPrefix(trimmedLine, "| ") ||
			trimmedLine == "##" ||
			trimmedLine == "###" ||
			trimmedLine == "####" ||
			trimmedLine == "**" ||
			trimmedLine == "***" ||
			strings.HasPrefix(trimmedLine, "-- ") ||
			strings.HasPrefix(trimmedLine, "-----") {
			skippedLineCount++
			continue // 跳过标记行
		}

		// 使用文本处理器进行详细预处理和验证
		if !cas.textProcessor.IsValidTextForTTS(line) {
			skippedLineCount++
			continue // 跳过无效行
		}

		// 处理文本以优化TTS效果
		processedText := cas.textProcessor.ProcessText(line)
		if processedText == "" {
			skippedLineCount++
			continue
		}

		validLineCount++
		tasks = append(tasks, TTSTask{Index: i, Text: processedText})
	}

	if len(tasks) == 0 {
		return fmt.Errorf("没有有效的文本行需要处理")
	}

	fmt.Printf("文本处理统计: 总行数=%d, 有效行数=%d, 跳过行数=%d\n", len(lines), validLineCount, skippedLineCount)

	// 并发处理任务
	results, err := cas.processTTSTasksConcurrent(tasks)
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

	// 提取音频文件路径
	audioFiles := make([]string, len(results))
	for i, result := range results {
		audioFiles[i] = result.AudioFile
	}

	// 合并音频文件
	return cas.mergeAudioFiles(audioFiles)
}

// processTTSTasksConcurrent 并发处理TTS任务
func (cas *ConcurrentAudioService) processTTSTasksConcurrent(tasks []TTSTask) ([]TTSResult, error) {
	ctx := context.Background()

	// 创建任务通道和结果通道
	taskChan := make(chan TTSTask, len(tasks))
	resultChan := make(chan TTSResult, len(tasks))

	// 发送所有任务到通道
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 启动worker goroutines
	var wg sync.WaitGroup
	numWorkers := cas.config.Concurrent.MaxWorkers
	if numWorkers > len(tasks) {
		numWorkers = len(tasks)
	}

	fmt.Printf("启动 %d 个worker开始处理...\n", numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			cas.worker(ctx, workerID, taskChan, resultChan)
		}(i)
	}

	// 等待所有worker完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var results []TTSResult
	successCount := 0
	failCount := 0

	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("任务 %d 失败: %v\n", result.Index, result.Error)
			failCount++
		} else {
			fmt.Printf("✓ 任务 %d 完成: %s\n", result.Index, result.AudioFile)
			results = append(results, result)
			successCount++
		}
	}

	fmt.Printf("\n处理完成: 成功 %d, 失败 %d\n", successCount, failCount)
	return results, nil
}

// worker 工作goroutine
func (cas *ConcurrentAudioService) worker(ctx context.Context, workerID int, taskChan <-chan TTSTask, resultChan chan<- TTSResult) {
	for task := range taskChan {
		// 等待速率限制
		if err := cas.limiter.Wait(ctx); err != nil {
			resultChan <- TTSResult{
				Index: task.Index,
				Error: fmt.Errorf("worker %d 等待速率限制失败: %v", workerID, err),
			}
			continue
		}

		fmt.Printf("Worker %d 处理任务 %d: %s\n", workerID, task.Index, task.Text)

		// 处理TTS任务
		audioFile, err := cas.generateAudioForText(task.Text, task.Index)

		resultChan <- TTSResult{
			Index:     task.Index,
			AudioFile: audioFile,
			Error:     err,
		}
	}
}

// readInputFile 读取历史文件
func (cas *ConcurrentAudioService) readInputFile() ([]string, error) {
	file, err := os.Open(cas.config.InputFile)
	if err != nil {
		return nil, fmt.Errorf("打开历史文件失败: %v", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取历史文件失败: %v", err)
	}

	return lines, nil
}

// generateAudioForText 为文本生成音频
func (cas *ConcurrentAudioService) generateAudioForText(text string, index int) (string, error) {
	// 创建TTS请求
	req := &model.TTSRequest{
		Text:            text,
		VoiceType:       cas.config.TTS.VoiceType,
		Volume:          cas.config.TTS.Volume,
		Speed:           cas.config.TTS.Speed,
		PrimaryLanguage: cas.config.TTS.PrimaryLanguage,
		SampleRate:      cas.config.TTS.SampleRate,
		Codec:           cas.config.TTS.Codec,
	}

	// 创建TTS任务
	resp, err := cas.ttsService.CreateTTSTask(req)
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", fmt.Errorf("创建TTS任务失败: %s", resp.Error)
	}

	// 等待任务完成并获取音频URL
	audioURL, err := cas.waitForTTSCompletion(resp.TaskID)
	if err != nil {
		return "", err
	}

	// 下载音频文件
	filename := fmt.Sprintf("audio_%03d.%s", index, cas.config.TTS.Codec)
	audioFile := filepath.Join(cas.config.Audio.TempDir, filename)

	err = cas.downloadAudio(audioURL, audioFile)
	if err != nil {
		return "", err
	}

	return audioFile, nil
}

// waitForTTSCompletion 等待TTS任务完成
func (cas *ConcurrentAudioService) waitForTTSCompletion(taskID string) (string, error) {
	maxRetries := 30 // 最多等待3分钟
	retryInterval := 6 * time.Second

	for i := 0; i < maxRetries; i++ {
		statusResp, err := cas.ttsService.DescribeTTSTaskStatus(taskID)
		if err != nil {
			return "", err
		}

		if !statusResp.Success {
			return "", fmt.Errorf("查询TTS任务状态失败: %s", statusResp.Error)
		}

		// 状态码：2表示成功
		if statusResp.Status == 2 {
			if statusResp.AudioURL == "" {
				return "", fmt.Errorf("TTS任务完成但未获取到音频URL")
			}
			return statusResp.AudioURL, nil
		}

		// 状态码：-1表示失败
		if statusResp.Status == -1 {
			return "", fmt.Errorf("TTS任务失败: %s", statusResp.ErrorMsg)
		}

		// 等待后重试
		time.Sleep(retryInterval)
	}

	return "", fmt.Errorf("TTS任务超时，任务ID: %s", taskID)
}

// downloadAudio 下载音频文件
func (cas *ConcurrentAudioService) downloadAudio(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载音频失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载音频失败，状态码: %d", resp.StatusCode)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建音频文件失败: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("保存音频文件失败: %v", err)
	}

	return nil
}

// mergeAudioFiles 合并音频文件
func (cas *ConcurrentAudioService) mergeAudioFiles(audioFiles []string) error {
	fmt.Printf("\n开始合并 %d 个音频文件...\n", len(audioFiles))

	outputPath := filepath.Join(cas.config.Audio.OutputDir, cas.config.Audio.FinalOutput)

	// 创建一个临时的文件列表
	listFile := filepath.Join(cas.config.Audio.TempDir, "file_list.txt")

	// 写入文件列表
	err := cas.createFileList(audioFiles, listFile)
	if err != nil {
		return err
	}
	defer os.Remove(listFile)

	// 使用简单合并
	return cas.simpleAudioMerge(listFile, outputPath)
}

// createFileList 创建文件列表
func (cas *ConcurrentAudioService) createFileList(audioFiles []string, listFile string) error {
	file, err := os.Create(listFile)
	if err != nil {
		return fmt.Errorf("创建文件列表失败: %v", err)
	}
	defer file.Close()

	for _, audioFile := range audioFiles {
		_, err := fmt.Fprintf(file, "file '%s'\n", audioFile)
		if err != nil {
			return fmt.Errorf("写入文件列表失败: %v", err)
		}
	}

	return nil
}

// simpleAudioMerge 简单的音频文件合并
func (cas *ConcurrentAudioService) simpleAudioMerge(listFile, outputPath string) error {
	// 读取文件列表
	listContent, err := os.ReadFile(listFile)
	if err != nil {
		return fmt.Errorf("读取文件列表失败: %v", err)
	}

	lines := strings.Split(string(listContent), "\n")
	var audioFiles []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 解析 "file 'path'" 格式
		if strings.HasPrefix(line, "file '") && strings.HasSuffix(line, "'") {
			filepath := line[6 : len(line)-1]
			audioFiles = append(audioFiles, filepath)
		}
	}

	if len(audioFiles) == 0 {
		return fmt.Errorf("没有找到要合并的音频文件")
	}

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outputFile.Close()

	// 按顺序合并音频文件
	for i, audioFile := range audioFiles {
		fmt.Printf("合并文件 %d/%d: %s\n", i+1, len(audioFiles), audioFile)

		inputFile, err := os.Open(audioFile)
		if err != nil {
			fmt.Printf("警告: 打开文件失败 %s: %v\n", audioFile, err)
			continue
		}

		_, err = io.Copy(outputFile, inputFile)
		inputFile.Close()

		if err != nil {
			fmt.Printf("警告: 复制文件失败 %s: %v\n", audioFile, err)
			continue
		}
	}

	fmt.Printf("音频合并完成: %s\n", outputPath)
	return nil
}

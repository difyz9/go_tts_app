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
	"text/tabwriter"
	"time"
	"tts_app/model"

	"github.com/difyz9/edge-tts-go/pkg/communicate"
	"github.com/difyz9/edge-tts-go/pkg/types"
	"github.com/difyz9/edge-tts-go/pkg/voices"
	"golang.org/x/time/rate"
)

// EdgeTTSTask Edge TTS任务结构
type EdgeTTSTask struct {
	Index int
	Text  string
}

// EdgeTTSResult Edge TTS任务结果
type EdgeTTSResult struct {
	Index     int
	AudioFile string
	Error     error
}

// EdgeTTSService Edge TTS服务
type EdgeTTSService struct {
	config  *model.Config
	limiter *rate.Limiter
}

// NewEdgeTTSService 创建Edge TTS服务
func NewEdgeTTSService(config *model.Config) *EdgeTTSService {
	// 创建速率限制器，Edge TTS可以更快一些
	rateLimit := rate.Every(time.Second / time.Duration(config.Concurrent.RateLimit))
	limiter := rate.NewLimiter(rateLimit, config.Concurrent.RateLimit)

	return &EdgeTTSService{
		config:  config,
		limiter: limiter,
	}
}

// ProcessInputFileConcurrent 并发处理输入文件
func (ets *EdgeTTSService) ProcessInputFileConcurrent() error {
	// 确保目录存在
	if err := os.MkdirAll(ets.config.Audio.TempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	if err := os.MkdirAll(ets.config.Audio.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 读取输入文件
	lines, err := ets.readInputFile()
	if err != nil {
		return err
	}

	fmt.Printf("读取到 %d 行文本，开始并发生成音频...\n", len(lines))
	fmt.Printf("并发配置: workers=%d, rate_limit=%d/秒, batch_size=%d\n",
		ets.config.Concurrent.MaxWorkers,
		ets.config.Concurrent.RateLimit,
		ets.config.Concurrent.BatchSize)

	// 创建任务列表
	tasks := make([]EdgeTTSTask, 0, len(lines))
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue // 跳过空行
		}

		// 跳过特定格式的标记行
		if strings.HasPrefix(trimmedLine, "###") ||
			strings.HasPrefix(trimmedLine, "**") ||
			strings.HasPrefix(trimmedLine, "-----") {
			continue // 跳过标记行
		}

		tasks = append(tasks, EdgeTTSTask{Index: i, Text: line})
	}

	if len(tasks) == 0 {
		return fmt.Errorf("没有有效的文本行需要处理")
	}

	fmt.Printf("跳过 %d 个空行/标记行，实际处理 %d 行有效文本\n", len(lines)-len(tasks), len(tasks))

	// 并发处理任务
	results, err := ets.processTTSTasksConcurrent(tasks)
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
	return ets.mergeAudioFiles(audioFiles)
}

// readInputFile 读取输入文件
func (ets *EdgeTTSService) readInputFile() ([]string, error) {
	file, err := os.Open(ets.config.InputFile)
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
func (ets *EdgeTTSService) processTTSTasksConcurrent(tasks []EdgeTTSTask) ([]EdgeTTSResult, error) {
	// 创建通道
	taskChan := make(chan EdgeTTSTask, len(tasks))
	resultChan := make(chan EdgeTTSResult, len(tasks))

	// 将任务发送到通道
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 确定worker数量
	workerCount := ets.config.Concurrent.MaxWorkers
	if workerCount > len(tasks) {
		workerCount = len(tasks)
	}

	fmt.Printf("启动 %d 个worker开始处理...\n", workerCount)

	// 启动workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go ets.edgeTTSWorker(i, taskChan, resultChan, &wg)
	}

	// 等待所有workers完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var results []EdgeTTSResult
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

// edgeTTSWorker Edge TTS工作协程
func (ets *EdgeTTSService) edgeTTSWorker(workerID int, taskChan <-chan EdgeTTSTask, resultChan chan<- EdgeTTSResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskChan {
		fmt.Printf("Worker %d 处理任务 %d: %s\n", workerID, task.Index, task.Text)

		// 限制请求频率
		err := ets.limiter.Wait(context.Background())
		if err != nil {
			resultChan <- EdgeTTSResult{
				Index: task.Index,
				Error: fmt.Errorf("等待速率限制失败: %v", err),
			}
			continue
		}

		// 生成音频
		audioFile, err := ets.generateAudioForText(task.Text, task.Index)
		resultChan <- EdgeTTSResult{
			Index:     task.Index,
			AudioFile: audioFile,
			Error:     err,
		}
	}
}

// generateAudioForText 为文本生成音频
func (ets *EdgeTTSService) generateAudioForText(text string, index int) (string, error) {
	ctx := context.Background()

	// 使用配置中的语音参数
	voice := ets.config.EdgeTTS.Voice
	if voice == "" {
		voice = "zh-CN-XiaoyiNeural" // 默认中文女声
	}

	rate := ets.config.EdgeTTS.Rate
	if rate == "" {
		rate = "+0%" // 默认正常语速
	}

	volume := ets.config.EdgeTTS.Volume
	if volume == "" {
		volume = "+0%" // 默认正常音量
	}

	pitch := ets.config.EdgeTTS.Pitch
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
	audioPath := filepath.Join(ets.config.Audio.TempDir, filename)

	// 保存音频文件
	err = comm.Save(ctx, audioPath, "")
	if err != nil {
		return "", fmt.Errorf("保存音频文件失败: %v", err)
	}

	return audioPath, nil
}

// mergeAudioFiles 合并音频文件
func (ets *EdgeTTSService) mergeAudioFiles(audioFiles []string) error {
	if len(audioFiles) == 0 {
		return fmt.Errorf("没有音频文件需要合并")
	}

	fmt.Printf("开始合并 %d 个音频文件...\n", len(audioFiles))

	// 输出文件路径
	outputPath := filepath.Join(ets.config.Audio.OutputDir, ets.config.Audio.FinalOutput)

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer outputFile.Close()

	// 逐个读取并合并音频文件
	for i, audioFile := range audioFiles {
		fmt.Printf("合并文件 %d/%d: %s\n", i+1, len(audioFiles), audioFile)

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

// ListEdgeVoices 列出可用的 Edge TTS 语音
func ListEdgeVoices(languageFilter string) error {
	ctx := context.Background()

	fmt.Println("正在获取Edge TTS语音列表...")

	// 获取语音列表
	voiceList, err := voices.ListVoices(ctx, "")
	if err != nil {
		return fmt.Errorf("获取语音列表失败: %v", err)
	}

	// 过滤语音（如果指定了语言）
	var filteredVoices []types.Voice
	if languageFilter != "" {
		languageFilter = strings.ToLower(languageFilter)
		for _, voice := range voiceList {
			// 检查语言代码（如 zh-CN, en-US, ja-JP）
			locale := strings.ToLower(voice.Locale)
			if strings.HasPrefix(locale, languageFilter) {
				filteredVoices = append(filteredVoices, voice)
			}
		}
		fmt.Printf("\n找到 %d 个 '%s' 语言的语音:\n\n", len(filteredVoices), languageFilter)
	} else {
		filteredVoices = voiceList
		fmt.Printf("\n找到 %d 个可用语音:\n\n", len(filteredVoices))
	}

	if len(filteredVoices) == 0 {
		return fmt.Errorf("没有找到匹配的语音")
	}

	// 简化显示：只显示简短名称和区域
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "音色\t区域")
	fmt.Fprintln(w, "--------\t--------")

	for _, voice := range filteredVoices {
		fmt.Fprintf(w, "%s\t%s\n", voice.ShortName, voice.Locale)
	}
	w.Flush()
	fmt.Println()

	// 显示使用示例
	if len(filteredVoices) > 0 {
		exampleVoice := filteredVoices[0].ShortName
		fmt.Printf("使用示例:\n")
		fmt.Printf("  # 使用 %s 语音\n", exampleVoice)
		fmt.Printf("  ./tts_app edge -i input.txt --voice %s\n", exampleVoice)
		fmt.Printf("  # 调整语速和音量\n")
		fmt.Printf("  ./tts_app edge -i input.txt --voice %s --rate +20%% --volume +10%%\n\n", exampleVoice)
	}

	return nil
}

// getLanguageName 根据语言代码返回语言名称
func getLanguageName(locale string) string {
	languageMap := map[string]string{
		"zh-CN": "中文(简体)",
		"zh-TW": "中文(繁体)",
		"zh-HK": "中文(香港)",
		"en-US": "英语(美国)",
		"en-GB": "英语(英国)",
		"en-AU": "英语(澳大利亚)",
		"en-CA": "英语(加拿大)",
		"ja-JP": "日语",
		"ko-KR": "韩语",
		"fr-FR": "法语",
		"de-DE": "德语",
		"es-ES": "西班牙语",
		"it-IT": "意大利语",
		"pt-BR": "葡萄牙语(巴西)",
		"ru-RU": "俄语",
		"ar-SA": "阿拉伯语",
		"hi-IN": "印地语",
		"th-TH": "泰语",
		"vi-VN": "越南语",
	}

	if name, exists := languageMap[locale]; exists {
		return name
	}
	return locale
}

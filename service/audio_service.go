package service

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tts_app/model"

	"gopkg.in/yaml.v3"
)

// ConfigService 配置服务
type ConfigService struct {
	config *model.Config
}

// NewConfigService 创建配置服务
func NewConfigService(configPath string) (*ConfigService, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &ConfigService{config: config}, nil
}

// GetConfig 获取配置
func (cs *ConfigService) GetConfig() *model.Config {
	return cs.config
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*model.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config model.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

// AudioMergeService 音频合并服务
type AudioMergeService struct {
	config     *model.Config
	ttsService *TTSService
}

// NewAudioMergeService 创建音频合并服务
func NewAudioMergeService(config *model.Config, ttsService *TTSService) *AudioMergeService {
	return &AudioMergeService{
		config:     config,
		ttsService: ttsService,
	}
}

// ProcessHistoryFile 处理历史文件，生成音频
func (ams *AudioMergeService) ProcessHistoryFile() error {
	// 确保目录存在
	if err := os.MkdirAll(ams.config.Audio.TempDir, 0755); err != nil {
		return fmt.Errorf("创建临时目录失败: %v", err)
	}
	if err := os.MkdirAll(ams.config.Audio.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 读取历史文件
	lines, err := ams.readHistoryFile()
	if err != nil {
		return err
	}

	fmt.Printf("读取到 %d 行文本，开始生成音频...\n", len(lines))

	// 为每行文本生成音频
	audioFiles := make([]string, 0, len(lines))
	skippedLines := 0
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			skippedLines++
			continue // 跳过空行
		}

		// 跳过特定格式的标记行
		if strings.HasPrefix(trimmedLine, "###") ||
			strings.HasPrefix(trimmedLine, "**") ||
			strings.HasPrefix(trimmedLine, "|") ||
			strings.HasPrefix(trimmedLine, "-----") {
			skippedLines++
			continue // 跳过标记行
		}

		fmt.Printf("正在处理第 %d 行: %s\n", i+1, line)
		audioFile, err := ams.generateAudioForText(line, i)
		if err != nil {
			fmt.Printf("生成第 %d 行音频失败: %v\n", i+1, err)
			continue
		}
		audioFiles = append(audioFiles, audioFile)
	}

	if len(audioFiles) == 0 {
		return fmt.Errorf("没有成功生成任何音频文件")
	}

	fmt.Printf("跳过 %d 个空行/标记行，成功处理 %d 行有效文本\n", skippedLines, len(audioFiles))

	// 合并音频文件
	return ams.mergeAudioFiles(audioFiles)
}

// readHistoryFile 读取历史文件
func (ams *AudioMergeService) readHistoryFile() ([]string, error) {
	file, err := os.Open(ams.config.InputFile)
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
func (ams *AudioMergeService) generateAudioForText(text string, index int) (string, error) {
	// 创建TTS请求
	req := &model.TTSRequest{
		Text:            text,
		VoiceType:       ams.config.TTS.VoiceType,
		Volume:          ams.config.TTS.Volume,
		Speed:           ams.config.TTS.Speed,
		PrimaryLanguage: ams.config.TTS.PrimaryLanguage,
		SampleRate:      ams.config.TTS.SampleRate,
		Codec:           ams.config.TTS.Codec,
	}

	// 创建TTS任务
	resp, err := ams.ttsService.CreateTTSTask(req)
	if err != nil {
		return "", err
	}

	if !resp.Success {
		return "", fmt.Errorf("创建TTS任务失败: %s", resp.Error)
	}

	// 等待任务完成并获取音频URL
	audioURL, err := ams.waitForTTSCompletion(resp.TaskID)
	if err != nil {
		return "", err
	}

	// 下载音频文件
	filename := fmt.Sprintf("audio_%03d.%s", index, ams.config.TTS.Codec)
	audioFile := filepath.Join(ams.config.Audio.TempDir, filename)

	err = ams.downloadAudio(audioURL, audioFile)
	if err != nil {
		return "", err
	}

	return audioFile, nil
}

// waitForTTSCompletion 等待TTS任务完成
func (ams *AudioMergeService) waitForTTSCompletion(taskID string) (string, error) {
	maxRetries := 30 // 最多等待3分钟（30次 * 6秒）
	retryInterval := 6 * time.Second

	for i := 0; i < maxRetries; i++ {
		statusResp, err := ams.ttsService.DescribeTTSTaskStatus(taskID)
		if err != nil {
			return "", err
		}

		if !statusResp.Success {
			return "", fmt.Errorf("查询TTS任务状态失败: %s", statusResp.Error)
		}

		fmt.Printf("TTS任务状态: %s\n", statusResp.StatusStr)

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
func (ams *AudioMergeService) downloadAudio(url, filepath string) error {
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

	fmt.Printf("音频文件已保存: %s\n", filepath)
	return nil
}

// mergeAudioFiles 合并音频文件
func (ams *AudioMergeService) mergeAudioFiles(audioFiles []string) error {
	fmt.Printf("开始合并 %d 个音频文件...\n", len(audioFiles))

	// 构建ffmpeg命令
	outputPath := filepath.Join(ams.config.Audio.OutputDir, ams.config.Audio.FinalOutput)

	// 创建一个临时的文件列表
	listFile := filepath.Join(ams.config.Audio.TempDir, "file_list.txt")

	// 写入文件列表
	err := ams.createFileList(audioFiles, listFile)
	if err != nil {
		return err
	}
	defer os.Remove(listFile) // 清理临时文件

	// 如果配置了静音间隔，使用复杂的合并方式
	if ams.config.Audio.SilenceDuration > 0 {
		return ams.mergeWithSilence(audioFiles, outputPath)
	}

	// 直接拼接音频文件
	return ams.concatAudioFiles(listFile, outputPath)
}

// createFileList 创建文件列表
func (ams *AudioMergeService) createFileList(audioFiles []string, listFile string) error {
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

// concatAudioFiles 直接拼接音频文件
func (ams *AudioMergeService) concatAudioFiles(listFile, outputPath string) error {
	// 检查ffmpeg是否可用
	if !ams.isFFmpegAvailable() {
		return ams.simpleAudioMerge(listFile, outputPath)
	}

	// 使用ffmpeg合并
	cmd := fmt.Sprintf("ffmpeg -f concat -safe 0 -i '%s' -c copy '%s' -y", listFile, outputPath)
	fmt.Printf("执行命令: %s\n", cmd)

	// 这里我们使用简单的文件合并作为备选方案
	return ams.simpleAudioMerge(listFile, outputPath)
}

// mergeWithSilence 带静音间隔的合并
func (ams *AudioMergeService) mergeWithSilence(audioFiles []string, outputPath string) error {
	if !ams.isFFmpegAvailable() {
		fmt.Println("警告: 未检测到ffmpeg，将使用简单拼接（无静音间隔）")
		listFile := filepath.Join(ams.config.Audio.TempDir, "file_list.txt")
		ams.createFileList(audioFiles, listFile)
		return ams.simpleAudioMerge(listFile, outputPath)
	}

	// 构建ffmpeg复杂过滤器命令
	var filterComplex strings.Builder
	var inputs strings.Builder

	for i, audioFile := range audioFiles {
		inputs.WriteString(fmt.Sprintf("-i '%s' ", audioFile))

		if i > 0 {
			// 添加静音
			silenceDuration := strconv.FormatFloat(ams.config.Audio.SilenceDuration, 'f', 1, 64)
			filterComplex.WriteString(fmt.Sprintf("[%d:0]adelay=%s[a%d]; ", i, silenceDuration+"s", i))
		}
	}

	// 添加音频混合
	filterComplex.WriteString("[0:0]")
	for i := 1; i < len(audioFiles); i++ {
		filterComplex.WriteString(fmt.Sprintf("[a%d]", i))
	}
	filterComplex.WriteString(fmt.Sprintf("concat=n=%d:v=0:a=1[out]", len(audioFiles)))

	cmd := fmt.Sprintf("ffmpeg %s -filter_complex '%s' -map '[out]' '%s' -y",
		inputs.String(), filterComplex.String(), outputPath)

	fmt.Printf("执行命令: %s\n", cmd)

	// 简化处理，直接使用简单合并
	listFile := filepath.Join(ams.config.Audio.TempDir, "file_list.txt")
	ams.createFileList(audioFiles, listFile)
	return ams.simpleAudioMerge(listFile, outputPath)
}

// isFFmpegAvailable 检查ffmpeg是否可用
func (ams *AudioMergeService) isFFmpegAvailable() bool {
	// 简单检查，实际项目中可以执行ffmpeg -version命令检查
	return false // 暂时返回false，使用简单合并
}

// simpleAudioMerge 简单的音频文件合并（二进制拼接）
func (ams *AudioMergeService) simpleAudioMerge(listFile, outputPath string) error {
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
			filepath := line[6 : len(line)-1] // 去掉 "file '" 和末尾的 "'"
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

	// 简单的二进制拼接（适用于相同格式的音频文件）
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
